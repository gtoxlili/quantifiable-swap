package swap

import (
	"errors"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common"
	"github.com/gtoxlili/quantifiable-swap/provider"
	"github.com/gtoxlili/quantifiable-swap/quantifiable"
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/tools/container/intsets"
	"strings"
	"time"
)

// LastTrade 最后一次 购买/卖出 的快照
type LastTrade struct {
	OrderTime time.Time
	Price     float64
	RSI       float64
}

// IndicatorWaper manages Indicator-based auto-trading logic
type IndicatorWaper struct {
	lastSellTrade *LastTrade
	lastBuyTrade  *LastTrade

	base  string
	quote string

	bar        time.Duration
	sellAmount string
	buyAmount  string
	autoTrade  bool

	// 判断是否自动下单的函数
	canSell func(tc TradeCondition) error
	canBuy  func(tc TradeCondition) error

	stopChan chan struct{}

	dataProvider provider.Provider
}

// NewNotify 不进行自动下单的 Waper （只提醒）
func NewNotify(base, quote string, bar time.Duration, dataProvider provider.Provider) *IndicatorWaper {
	return NewIndicatorWaperWithCustomSellBuy(base, quote, bar, "", "", false, nil, nil, dataProvider)
}

// NewWaper creates a new Waper instance
func NewWaper(base, quote string, bar time.Duration, sellAmount, buyAmount string, dataProvider provider.Provider) *IndicatorWaper {
	return NewIndicatorWaperWithCustomSellBuy(base, quote, bar, sellAmount, buyAmount, true, defaultCanSell, defaultCanBuy, dataProvider)
}

func NewIndicatorWaperWithCustomSellBuy(base, quote string, bar time.Duration, sellAmount, buyAmount string, autoTrade bool, canSell, canBuy func(tc TradeCondition) error, dataProvider provider.Provider) *IndicatorWaper {
	return &IndicatorWaper{
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

		// 初始化快照
		lastSellTrade: &LastTrade{
			OrderTime: time.Unix(0, 0),
			RSI:       float64(intsets.MaxInt),
			Price:     float64(intsets.MaxInt),
		},
		lastBuyTrade: &LastTrade{
			OrderTime: time.Unix(0, 0),
			RSI:       float64(intsets.MinInt),
			Price:     float64(intsets.MinInt),
		},
	}
}

func (r *IndicatorWaper) Stop() {
	close(r.stopChan)
}

func (r *IndicatorWaper) Run() {
	r.RunWithCustomPeriod(14)
}

func (r *IndicatorWaper) RunWithCustomPeriod(period int) {
	rsiHook, err := r.prepareIndicatorHook(period)
	if err != nil {
		fmt.Printf("[%s] %v\n", r.printInstId(), err)
		return
	}
	r.runIndicatorLoop(rsiHook)
}

func (r *IndicatorWaper) prepareIndicatorHook(period int) (quantifiable.IndicatorDecorator[float64], error) {
	priceSeq, err := sequence.NewPriceSequence(r.base, r.quote, r.bar, common.ExtraPointsForInitialDecay(period), r.dataProvider)
	if err != nil {
		return nil, fmt.Errorf("创建价格序列失败：%w", err)
	}

	return quantifiable.NewIndicatorBuilder(priceSeq).
		WithMA(5).
		WithMA(20).
		WithRSI(period).
		Build()
}

func (r *IndicatorWaper) runIndicatorLoop(hook quantifiable.IndicatorDecorator[float64]) {
	for {
		select {
		case <-r.stopChan:
			fmt.Printf("[%s] RSIWAP 策略已停止\n", r.printInstId())
			return
		default:
			candle, err := hook.Update()
			if err != nil {
				fmt.Printf("更新价格序列失败：%v\n", err)
				return
			}

			rsiHook, _ := hook.Indicator("RSI")
			ma5Hook, _ := hook.Indicator("MA5")
			ma20Hook, _ := hook.Indicator("MA20")

			curRSI := rsiHook.CurrentVal()
			ma5 := ma5Hook.CurrentVal()
			ma20 := ma20Hook.CurrentVal()
			lstMA5 := ma5Hook.PreviousVals()
			lstMA20 := ma20Hook.PreviousVals()

			// 当前价格，当前 RSI
			fmt.Printf("[%s][%s][%s]: Time: %s, Price: %.2f, RSI: %.2f, MA5: %.2f, MA20: %.2f\n",
				r.printInstId(),
				r.dataProvider.Name(),
				fmt.Sprintf("%dm", int(r.bar.Minutes())),
				candle.Time.Format("15:04:05"),
				candle.Value, curRSI,
				ma5, ma20,
			)

			// 交叉提醒
			if text := crossOver(ma5, ma20, lstMA5[4], lstMA20[4]); text != "" {
				common.Notify(
					"⚠️ ["+r.printInstId()+"] 均线交叉提醒",
					fmt.Sprintf("[%s][%s]%s Price: %.2f, MA5: %.2f, MA20: %.2f",
						r.dataProvider.Name(),
						fmt.Sprintf("%dm", int(r.bar.Minutes())), text,
						candle.Value, ma5, ma20),
					"指标监控",
				)
			}

			// 通用 RSI 提醒
			if curRSI < 30 || curRSI > 70 {
				common.Notify(
					"⚠️ ["+r.printInstId()+"] RSI提醒",
					fmt.Sprintf("[%s][%s] Price: %.2f, RSI: %.2f",
						r.dataProvider.Name(),
						fmt.Sprintf("%dm", int(r.bar.Minutes())),
						candle.Value, curRSI),
					"指标监控",
				)
			}

			// 如果自动交易未开启，直接继续循环
			if !r.autoTrade {
				continue
			}

			condition := TradeCondition{
				rsiQueue: rsiHook.PreviousVals(),
				curRSI:   curRSI,
				lst:      r.lastBuyTrade,
				bar:      r.bar,
				price:    candle.Value,
				ma5:      ma5,
				ma20:     ma20,
			}

			if err := r.canBuy(condition); err != nil {
				if !errors.Is(err, ErrNotMeetBuyCondition) && !errors.Is(err, ErrInsufficientSampleData) {
					fmt.Printf("[%s] Time: %s, %v\n", r.printInstId(), candle.Time.Format("15:04:05"), err)
				}
			} else {
				orderID, err := r.dataProvider.MarketOrder(r.base, r.quote, "buy", r.buyAmount)
				if err != nil {
					fmt.Printf("[%s]买入失败：%v\n", r.printInstId(), err)
					common.Notify(
						"❌ ["+r.printInstId()+"]自动买入失败",
						fmt.Sprintf("Error: %v", err),
						"自动交易",
					)
				} else {
					common.Notify(
						"🚀 ["+r.printInstId()+"]自动买入提醒",
						fmt.Sprintf("Price: %.2f, RSI: %.2f, OrderId: %s", candle.Value, curRSI, orderID),
						"自动交易",
					)
					fmt.Printf("[%s]买入成功，订单号：%s\n", r.printInstId(), orderID)
					r.lastBuyTrade.OrderTime = time.Now()
					r.lastBuyTrade.Price = candle.Value
					r.lastBuyTrade.RSI = curRSI
				}
			}

			if err := r.canSell(condition); err != nil {
				if !errors.Is(err, ErrInsufficientSampleData) && !errors.Is(err, ErrNotMeetSellCondition) {
					fmt.Printf("[%s] Time: %s, %v\n", r.printInstId(), candle.Time.Format("15:04:05"), err)
				}
			} else {
				orderID, err := r.dataProvider.MarketOrder(r.base, r.quote, "sell", r.sellAmount)
				if err != nil {
					fmt.Printf("[%s]卖出失败：%v\n", r.printInstId(), err)
					common.Notify(
						"❌ ["+r.printInstId()+"]自动卖出失败",
						fmt.Sprintf("Error: %v", err),
						"自动交易",
					)
				} else {
					common.Notify(
						"🚀 ["+r.printInstId()+"]自动卖出提醒",
						fmt.Sprintf("Price: %.2f, RSI: %.2f, OrderId: %s", candle.Value, curRSI, orderID),
						"自动交易",
					)
					fmt.Printf("[%s]卖出成功，订单号：%s\n", r.printInstId(), orderID)
					r.lastSellTrade.OrderTime = time.Now()
					r.lastSellTrade.Price = candle.Value
					r.lastSellTrade.RSI = curRSI
				}
			}
		}
	}
}

// 美化打印交易对
func (r *IndicatorWaper) printInstId() string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(r.base), strings.ToUpper(r.quote))
}
