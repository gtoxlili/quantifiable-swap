package trading

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/common/lo"
	"github.com/gtoxlili/quantifiable-swap/indicator"
	"github.com/gtoxlili/quantifiable-swap/logger"
	"github.com/gtoxlili/quantifiable-swap/market"
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/tools/container/intsets"
	"os"
	"strings"
	"time"
)

// IStrategyExecutor 接口
type IStrategyExecutor interface {
	Run(ctx context.Context)
	WithSubscribers(subscribers []config.Subscriber)
	RunWithCustomPeriod(ctx context.Context, period int)
}

// TradeSnapshot 最后一次 购买/卖出 的快照
type TradeSnapshot struct {
	OrderTime time.Time `json:"order_time"`
	Price     float64   `json:"price"`
	RSI       float64   `json:"rsi"`
}

// StrategyExecutor manages Indicator-based auto-trading logic
type StrategyExecutor struct {
	lastSellSnapshot *TradeSnapshot
	lastBuySnapshot  *TradeSnapshot

	base       string
	quote      string
	bar        time.Duration
	sellAmount float64
	buyAmount  float64
	autoTrade  bool

	// 判断是否自动下单的函数
	sellStrategy func(tc TradeContext) error
	buyStrategy  func(tc TradeContext) error

	dataProvider    market.DataProvider
	tradingProvider market.TradingProvider

	log *logger.Logger
}

// NewMonitor 不进行自动下单的 executor （只提醒）
func NewMonitor(base, quote string, bar time.Duration, dataProvider market.DataProvider) IStrategyExecutor {
	return NewStrategyExecutorWithCustomStrategies(base, quote, bar, 0, 0, false, nil, nil, dataProvider, nil)
}

// NewTrader creates a new executor instance
func NewTrader(
	base, quote string,
	bar time.Duration,
	sellAmount, buyAmount float64,
	dataProvider market.DataProvider,
	tradingProvider market.TradingProvider,
) IStrategyExecutor {
	return NewStrategyExecutorWithCustomStrategies(base, quote, bar, sellAmount, buyAmount, true, defaultSellStrategy, defaultBuyStrategy, dataProvider, tradingProvider)
}

func NewStrategyExecutorWithCustomStrategies(
	base, quote string,
	bar time.Duration,
	sellAmount, buyAmount float64,
	autoTrade bool,
	sellStrategy, buyStrategy func(tc TradeContext) error,
	dataProvider market.DataProvider,
	tradingProvider market.TradingProvider,
) IStrategyExecutor {
	ind := &StrategyExecutor{
		base:            base,
		quote:           quote,
		bar:             bar,
		sellAmount:      sellAmount,
		buyAmount:       buyAmount,
		autoTrade:       autoTrade,
		sellStrategy:    sellStrategy,
		buyStrategy:     buyStrategy,
		dataProvider:    dataProvider,
		tradingProvider: tradingProvider,
	}
	urlScheme := dataProvider.UrlScheme(base, quote)
	if tradingProvider != nil {
		urlScheme = tradingProvider.UrlScheme(base, quote)
	}
	ind.log = logger.NewTraderLogger(ind.printInstId(), ind.dataProvider.Name(), int(ind.bar.Minutes()), urlScheme)
	ind.lastBuySnapshot = ind.loadSnapshot("buy")
	ind.lastSellSnapshot = ind.loadSnapshot("sell")
	return ind
}

func (e *StrategyExecutor) WithSubscribers(subscribers []config.Subscriber) {
	e.log = e.log.WithSubscribers(subscribers)
}

func (e *StrategyExecutor) Run(ctx context.Context) {
	e.RunWithCustomPeriod(ctx, 14)
}

func (e *StrategyExecutor) RunWithCustomPeriod(ctx context.Context, period int) {
	rsiHook, err := e.prepareStrategyHook(ctx, period)
	if err != nil {
		e.log.PrintError(err, false)
		return
	}
	e.executeStrategyLoop(ctx, rsiHook)
}

func (e *StrategyExecutor) prepareStrategyHook(ctx context.Context, period int) (indicator.Decorator[float64], error) {
	priceSeq, err := sequence.NewPriceSequence(ctx, e.base, e.quote, e.bar, common.ExtraPointsForInitialDecay(period), e.dataProvider)
	if err != nil {
		return nil, fmt.Errorf("创建价格序列失败：%w", err)
	}

	return indicator.NewIndicatorBuilder(priceSeq).
		WithMA(5).
		WithMA(20).
		WithRSI(period).
		Build()
}

func (e *StrategyExecutor) executeStrategyLoop(ctx context.Context, hook indicator.Decorator[float64]) {
	for {
		select {
		case <-ctx.Done():
			e.log.PrintStrategyStop()
			return
		default:
			candle, err := hook.Update(ctx)
			if err != nil {
				e.log.PrintError(fmt.Errorf("更新价格序列失败：%w", err), false)
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

			abnormal := ""
			// 交叉提醒
			if candle.Time.Truncate(e.bar).Equal(candle.Time) {
				if text := crossOver(ma5, ma20, lstMA5[4], lstMA20[4]); text != "" {
					abnormal = text
				}
			}
			// 通用 RSI 提醒
			if curRSI < 30 || curRSI > 70 {
				abnormal = "RSI"
			}
			e.log.PrintStrategyMetrics(candle.Time, candle.Value, curRSI, ma5, ma20, abnormal)

			// 如果自动交易未开启，直接继续循环
			if !e.autoTrade {
				continue
			}

			condition := TradeContext{
				rsiQueue: rsiHook.PreviousVals(),
				curRSI:   curRSI,
				snapshot: nil,
				bar:      e.bar,
				price:    candle.Value,
				ma5:      ma5,
				ma20:     ma20,
			}

			if err := e.buyStrategy(condition.Snapshot(e.lastBuySnapshot)); err != nil {
				if !errors.Is(err, ErrInvalidBuyCondition) && !errors.Is(err, ErrInsufficientData) {
					e.log.PrintErrorWithTime(candle.Time, err, true)
				}
			} else {
				orderID, err := e.tradingProvider.ExecuteMarketOrder(e.base, e.quote, "buy", e.buyAmount)
				if err != nil {
					e.log.PrintBuyFail(err)
				} else {
					e.log.PrintBuySuccess(candle.Value, curRSI, orderID)
					e.saveSnapshot("buy", candle.Value, curRSI)
				}
			}

			if err := e.sellStrategy(condition.Snapshot(e.lastSellSnapshot)); err != nil {
				if !errors.Is(err, ErrInsufficientData) && !errors.Is(err, ErrInvalidSellCondition) {
					e.log.PrintErrorWithTime(candle.Time, err, true)
				}
			} else {
				orderID, err := e.tradingProvider.ExecuteMarketOrder(e.base, e.quote, "sell", e.sellAmount)
				if err != nil {
					e.log.PrintSellFail(err)
				} else {
					e.log.PrintSellSuccess(candle.Value, curRSI, orderID)
					e.saveSnapshot("sell", candle.Value, curRSI)
				}
			}
		}
	}
}

// 美化打印交易对
func (e *StrategyExecutor) printInstId() string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(e.base), strings.ToUpper(e.quote))
}

// 保存快照
func (e *StrategyExecutor) saveSnapshot(typ string, price, curRSI float64) {
	snapshot := lo.IfThen(typ == "sell", e.lastSellSnapshot, e.lastBuySnapshot)
	snapshot.OrderTime = time.Now()
	snapshot.Price = price
	snapshot.RSI = curRSI

	if err := e.persistSnapshot(typ, snapshot); err != nil {
		e.log.PrintError(fmt.Errorf("持久化「%s」快照失败：%w", typ, err), false)
	}
}

func (e *StrategyExecutor) persistSnapshot(typ string, snapshot *TradeSnapshot) error {
	file, err := os.OpenFile(fmt.Sprintf("snapshot_%s_%s_%s_%d", typ, e.printInstId(), e.dataProvider.Name(), int(e.bar.Minutes())), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open snapshot file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(snapshot); err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	return nil
}

func (e *StrategyExecutor) loadSnapshot(typ string) *TradeSnapshot {
	defaultSnapshot := &TradeSnapshot{
		OrderTime: time.Unix(0, 0),
		RSI:       lo.IfThen(typ == "sell", float64(intsets.MaxInt), float64(intsets.MinInt)),
		Price:     lo.IfThen(typ == "sell", float64(intsets.MaxInt), float64(intsets.MinInt)),
	}

	file, err := os.Open(fmt.Sprintf("snapshot_%s_%s_%s_%d", typ, e.printInstId(), e.dataProvider.Name(), int(e.bar.Minutes())))
	if err != nil {
		if !os.IsNotExist(err) {
			e.log.PrintError(fmt.Errorf("打开「%s」快照失败：%w", typ, err), false)
		}
		return defaultSnapshot
	}
	defer file.Close()

	var snapshot TradeSnapshot
	if err := json.NewDecoder(file).Decode(&snapshot); err != nil {
		e.log.PrintError(fmt.Errorf("解析「%s」快照失败：%w", typ, err), false)
	}
	return &snapshot
}
