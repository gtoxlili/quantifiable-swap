package trading

import (
	"context"
	"errors"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common"
	"github.com/gtoxlili/quantifiable-swap/indicator"
	"github.com/gtoxlili/quantifiable-swap/logger"
	"github.com/gtoxlili/quantifiable-swap/market"
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/tools/container/intsets"
	"strings"
	"time"
)

// IStrategyExecutor 接口
type IStrategyExecutor interface {
	Run(ctx context.Context)
	WithSubscribers(subscribers []int64)
	RunWithCustomPeriod(ctx context.Context, period int)
}

// TradeSnapshot 最后一次 购买/卖出 的快照
type TradeSnapshot struct {
	OrderTime time.Time
	Price     float64
	RSI       float64
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

	dataProvider market.Provider
	log          *logger.Logger
}

// NewMonitor 不进行自动下单的 executor （只提醒）
func NewMonitor(base, quote string, bar time.Duration, dataProvider market.Provider) IStrategyExecutor {
	return NewStrategyExecutorWithCustomStrategies(base, quote, bar, 0, 0, false, nil, nil, dataProvider)
}

// NewTrader creates a new executor instance
func NewTrader(base, quote string, bar time.Duration, sellAmount, buyAmount float64, dataProvider market.Provider) IStrategyExecutor {
	return NewStrategyExecutorWithCustomStrategies(base, quote, bar, sellAmount, buyAmount, true, defaultSellStrategy, defaultBuyStrategy, dataProvider)
}

func NewStrategyExecutorWithCustomStrategies(base, quote string, bar time.Duration, sellAmount, buyAmount float64, autoTrade bool, sellStrategy, buyStrategy func(tc TradeContext) error, dataProvider market.Provider) IStrategyExecutor {
	ind := &StrategyExecutor{
		base:         base,
		quote:        quote,
		bar:          bar,
		sellAmount:   sellAmount,
		buyAmount:    buyAmount,
		autoTrade:    autoTrade,
		sellStrategy: sellStrategy,
		buyStrategy:  buyStrategy,
		dataProvider: dataProvider,

		// 初始化快照
		lastSellSnapshot: &TradeSnapshot{
			OrderTime: time.Unix(0, 0),
			RSI:       float64(intsets.MaxInt),
			Price:     float64(intsets.MaxInt),
		},
		lastBuySnapshot: &TradeSnapshot{
			OrderTime: time.Unix(0, 0),
			RSI:       float64(intsets.MinInt),
			Price:     float64(intsets.MinInt),
		},
	}
	ind.log = logger.NewTraderLogger(ind.printInstId(), ind.dataProvider.Name(), int(ind.bar.Minutes()))
	return ind
}

func (r *StrategyExecutor) WithSubscribers(subscribers []int64) {
	r.log = r.log.WithSubscribers(subscribers)
}

func (r *StrategyExecutor) Run(ctx context.Context) {
	r.RunWithCustomPeriod(ctx, 14)
}

func (r *StrategyExecutor) RunWithCustomPeriod(ctx context.Context, period int) {
	rsiHook, err := r.prepareStrategyHook(ctx, period)
	if err != nil {
		r.log.PrintError(err, false)
		return
	}
	r.executeStrategyLoop(ctx, rsiHook)
}

func (r *StrategyExecutor) prepareStrategyHook(ctx context.Context, period int) (indicator.Decorator[float64], error) {
	priceSeq, err := sequence.NewPriceSequence(ctx, r.base, r.quote, r.bar, common.ExtraPointsForInitialDecay(period), r.dataProvider)
	if err != nil {
		return nil, fmt.Errorf("创建价格序列失败：%w", err)
	}

	return indicator.NewIndicatorBuilder(priceSeq).
		WithMA(5).
		WithMA(20).
		WithRSI(period).
		Build()
}

func (r *StrategyExecutor) executeStrategyLoop(ctx context.Context, hook indicator.Decorator[float64]) {
	for {
		select {
		case <-ctx.Done():
			r.log.PrintStrategyStop()
			return
		default:
			candle, err := hook.Update(ctx)
			if err != nil {
				r.log.PrintError(fmt.Errorf("更新价格序列失败：%w", err), false)
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
			if candle.Time.Truncate(r.bar).Equal(candle.Time) {
				if text := crossOver(ma5, ma20, lstMA5[4], lstMA20[4]); text != "" {
					abnormal = text
				}
			}
			// 通用 RSI 提醒
			if curRSI < 30 || curRSI > 70 {
				abnormal = "RSI"
			}
			r.log.PrintStrategyMetrics(candle.Time, candle.Value, curRSI, ma5, ma20, abnormal)

			// 如果自动交易未开启，直接继续循环
			if !r.autoTrade {
				continue
			}

			condition := TradeContext{
				rsiQueue: rsiHook.PreviousVals(),
				curRSI:   curRSI,
				snapshot: nil,
				bar:      r.bar,
				price:    candle.Value,
				ma5:      ma5,
				ma20:     ma20,
			}

			if err := r.buyStrategy(condition.Snapshot(r.lastBuySnapshot)); err != nil {
				if !errors.Is(err, ErrInvalidBuyCondition) && !errors.Is(err, ErrInsufficientData) {
					r.log.PrintErrorWithTime(candle.Time, err, true)
				}
			} else {
				orderID, err := r.dataProvider.ExecuteMarketOrder(r.base, r.quote, "buy", r.buyAmount)
				if err != nil {
					r.log.PrintBuyFail(err)
				} else {
					r.log.PrintBuySuccess(candle.Value, curRSI, orderID)
					r.lastBuySnapshot.OrderTime = time.Now()
					r.lastBuySnapshot.Price = candle.Value
					r.lastBuySnapshot.RSI = curRSI
				}
			}

			if err := r.sellStrategy(condition.Snapshot(r.lastSellSnapshot)); err != nil {
				if !errors.Is(err, ErrInsufficientData) && !errors.Is(err, ErrInvalidSellCondition) {
					r.log.PrintErrorWithTime(candle.Time, err, true)
				}
			} else {
				orderID, err := r.dataProvider.ExecuteMarketOrder(r.base, r.quote, "sell", r.sellAmount)
				if err != nil {
					r.log.PrintSellFail(err)
				} else {
					r.log.PrintSellSuccess(candle.Value, curRSI, orderID)
					r.lastSellSnapshot.OrderTime = time.Now()
					r.lastSellSnapshot.Price = candle.Value
					r.lastSellSnapshot.RSI = curRSI
				}
			}
		}
	}
}

// 美化打印交易对
func (r *StrategyExecutor) printInstId() string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(r.base), strings.ToUpper(r.quote))
}
