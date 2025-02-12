package swap

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common"
	"github.com/gtoxlili/quantifiable-swap/provider"
	"github.com/gtoxlili/quantifiable-swap/quantifiable"
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/exp/slices"
	"strings"
	"time"
)

const (
	// 价格序列的最大长度，超过这个长度则删除最旧的数据
	maxPriceCount = 300
)

// 最后一次 购买/卖出 的快照
type lastTrade struct {
	lastOrderTime time.Time
	lastPrice     float64
	lastRSI       float64
}

// RSIWaper manages RSI-based auto-trading logic
type RSIWaper struct {
	lastSellTrade *lastTrade
	lastBuyTrade  *lastTrade

	base  string
	quote string

	bar        time.Duration
	sellAmount string
	buyAmount  string
	autoTrade  bool

	// 判断是否自动下单的函数
	canSell func(rsiQueue []float64, curRSI float64) bool
	canBuy  func(rsiQueue []float64, curRSI float64) bool

	stopChan chan struct{}

	dataProvider provider.Provider
}

// NewRSINotify 不进行自动下单的 Waper （只提醒）
func NewRSINotify(base, quote string, bar time.Duration, dataProvider provider.Provider) *RSIWaper {
	return NewRSIWaperWithCustomSellBuy(base, quote, bar, "", "", false, nil, nil, dataProvider)
}

// NewRSIWaper creates a new RSIWaper instance
func NewRSIWaper(base, quote string, bar time.Duration, sellAmount, buyAmount string, dataProvider provider.Provider) *RSIWaper {
	return NewRSIWaperWithCustomSellBuy(base, quote, bar, sellAmount, buyAmount, true, defaultCanSell, defaultCanBuy, dataProvider)
}

func NewRSIWaperWithCustomSellBuy(base, quote string, bar time.Duration, sellAmount, buyAmount string, autoTrade bool, canSell, canBuy func(rsiQueue []float64, curRSI float64) bool, dataProvider provider.Provider) *RSIWaper {
	return &RSIWaper{
		base:         base,
		quote:        quote,
		bar:          bar,
		sellAmount:   sellAmount,
		buyAmount:    buyAmount,
		autoTrade:    autoTrade,
		canSell:      canSell,
		canBuy:       canBuy,
		stopChan:     make(chan struct{}),
		dataProvider: dataProvider,
	}
}

func (r *RSIWaper) Stop() {
	close(r.stopChan)
}

// Run starts the auto-trading routine based on RSI signals
func (r *RSIWaper) Run() {
	// 创建一个价格序列，保存最多 maxPriceCount 个数据
	priceSeq, err := sequence.NewPriceSequence(r.base, r.quote, r.bar, maxPriceCount, r.dataProvider)
	if err != nil {
		fmt.Printf("创建价格序列失败：%v\n", err)
		return
	}

	// 创建一个 RSI 包装器，计算周期为 14
	rsiHook, err := quantifiable.NewRSIHOOK(14, priceSeq)
	if err != nil {
		fmt.Printf("创建 RSI 包装器失败：%v\n", err)
		return
	}

	for {
		select {
		case <-r.stopChan:
			fmt.Printf("[%s] RSIWAP 策略已停止\n", r.printInstId())
			return
		default:
			candle, err := rsiHook.Update()
			if err != nil {
				fmt.Printf("更新价格序列失败：%v\n", err)
				continue
			}

			rsiQueue := rsiHook.LastRSIs()
			curRSI := rsiHook.CurrentRSI()

			// 当前价格，当前 RSI
			fmt.Printf("[%s][%s][%s]: Time: %s, Price: %.2f, RSI: %.2f\n", r.printInstId(), r.dataProvider.Name(), fmt.Sprintf("%dm", int(r.bar.Minutes())), candle.Time.Format("15:04:05"), candle.Value, curRSI)

			// 通用 RSI 提醒
			if curRSI < 30 || curRSI > 70 {
				common.Notify(
					"⚠️ ["+r.printInstId()+"]RSI提醒",
					fmt.Sprintf("[%s][%s] Price: %.2f, RSI: %.2f",
						r.dataProvider.Name(),
						fmt.Sprintf("%dm", int(r.bar.Minutes())),
						candle.Value, curRSI),
					false,
				)
			}

			// 如果自动交易未开启，直接继续循环
			if !r.autoTrade {
				continue
			}

			// 判断是否满足卖出条件
			if r.canSell(rsiQueue, curRSI) {
				if time.Since(r.lastSellTrade.lastOrderTime) <= 4*r.bar && curRSI < 1.2*r.lastSellTrade.lastRSI {
					// 避免短期一直在 70 附近震荡的情况
					r.lastSellTrade.lastOrderTime = time.Now()
					fmt.Printf("[%s]卖出太频繁 Time: %s\n", r.printInstId(), time.Now().Format("15:04:05"))
					continue
				}
				orderID, err := r.dataProvider.MarketOrder(r.base, r.quote, "sell", r.sellAmount)
				if err != nil {
					fmt.Printf("[%s]卖出失败：%v\n", r.printInstId(), err)
					common.Notify(
						"❌ ["+r.printInstId()+"]自动卖出失败",
						fmt.Sprintf("Error: %v", err),
						false,
					)
				} else {
					common.Notify(
						"🚀 ["+r.printInstId()+"]自动卖出提醒",
						fmt.Sprintf("Price: %.2f, RSI: %.2f, OrderId: %s", candle.Value, curRSI, orderID),
						false,
					)
					fmt.Printf("[%s]卖出成功，订单号：%s\n", r.printInstId(), orderID)
					r.lastSellTrade = &lastTrade{
						lastOrderTime: time.Now(),
						lastPrice:     candle.Value,
						lastRSI:       curRSI,
					}
				}
			} else if r.canBuy(rsiQueue, curRSI) {
				// 比如上次买的时候是 30
				// 如果这次 RSI 一直在 30 附近，就不买了
				// 但如果 跌回 20 以下，再次买入
				// 20 >= (2*30)/3
				if time.Since(r.lastBuyTrade.lastOrderTime) <= 4*r.bar && 1.2*curRSI > r.lastBuyTrade.lastRSI {
					// 避免短期一直在 30 附近震荡的情况
					r.lastBuyTrade.lastOrderTime = time.Now()
					fmt.Printf("[%s]买入太频繁\n", r.printInstId())
					continue
				}
				orderID, err := r.dataProvider.MarketOrder(r.base, r.quote, "buy", r.buyAmount)
				if err != nil {
					fmt.Printf("[%s]买入失败：%v\n", r.printInstId(), err)
					common.Notify(
						"❌ ["+r.printInstId()+"]自动买入失败",
						fmt.Sprintf("Error: %v", err),
						false,
					)
				} else {
					common.Notify(
						"🚀 ["+r.printInstId()+"]自动买入提醒",
						fmt.Sprintf("Price: %.2f, RSI: %.2f, OrderId: %s", candle.Value, curRSI, orderID),
						false,
					)
					fmt.Printf("[%s]买入成功，订单号：%s\n", r.printInstId(), orderID)
					r.lastBuyTrade = &lastTrade{
						lastOrderTime: time.Now(),
						lastPrice:     candle.Value,
						lastRSI:       curRSI,
					}
				}
			}
		}
	}
}

// 美化打印交易对
func (r *RSIWaper) printInstId() string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(r.base), strings.ToUpper(r.quote))
}

// canSell determines if the RSI suggests a sell signal
func defaultCanSell(rsiQueue []float64, curRSI float64) bool {
	// 如果包含 -1，说明数据不足，无法判断
	if slices.Contains(rsiQueue, -1) {
		return false
	}
	// 下标顺序: [0,1,2,3,4] -> older -> newer
	lastRSI := rsiQueue[4]
	lastLastRSI := rsiQueue[3]
	lastLastLastRSI := rsiQueue[2]

	// 卖出判断逻辑: up -> up -> * -> down && RSI > 70
	return lastRSI > lastLastRSI && lastLastRSI > lastLastLastRSI &&
		curRSI < lastRSI && lastRSI > 70
}

// canBuy determines if the RSI suggests a buy signal
func defaultCanBuy(rsiQueue []float64, curRSI float64) bool {
	// 如果包含 -1，说明数据不足，无法判断
	if slices.Contains(rsiQueue, -1) {
		return false
	}
	// 下标顺序: [0,1,2,3,4] -> older -> newer
	lastRSI := rsiQueue[4]
	lastLastRSI := rsiQueue[3]
	lastLastLastRSI := rsiQueue[2]

	// 买入判断逻辑: down -> down -> * -> up && RSI < 30
	return lastRSI < lastLastRSI && lastLastRSI < lastLastLastRSI &&
		curRSI > lastRSI && lastRSI < 30
}
