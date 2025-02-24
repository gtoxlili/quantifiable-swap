package logger

import (
	"context"
	"errors"
	"github.com/gtoxlili/quantifiable-swap/bot"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"github.com/gtoxlili/quantifiable-swap/logger/pretty/bark"
	"github.com/gtoxlili/quantifiable-swap/logger/pretty/console"
	"github.com/gtoxlili/quantifiable-swap/logger/pretty/tglog"
	"github.com/rs/zerolog"
	"io"
	"os"
	"time"
)

// Logger 是一个使用 zerolog 封装日志记录逻辑的 logger。
type Logger struct {
	log zerolog.Logger
}

var (
	l *Logger
)

func init() {
	zerolog.TimeFieldFormat = "15:04:05"
	lever, err := zerolog.ParseLevel(constants.LogLevel)
	if err != nil {
		lever = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lever)
	zerolog.FloatingPointPrecision = 2

	var outSet []io.Writer
	outSet = append(outSet, console.NewConsoleWriter(os.Stdout, "15:04:05"))

	if bot.Bot != nil {
		outSet = append(outSet, tglog.NewBotWriter(bot.Bot))
	}

	if constants.BarkToken != "" {
		outSet = append(outSet, bark.NewNotifyWriter(constants.BarkToken))
	}

	l = &Logger{
		log: zerolog.New(zerolog.MultiLevelWriter(outSet...)).With().Timestamp().Logger(),
	}
}

// NewTraderLogger 创建一个新的 Logger 单例实例。
func NewTraderLogger(instID, dataProvider string, barMinutes int, urlScheme string) *Logger {
	return &Logger{
		log: l.log.With().
			Str("type", "trader").       // 日志类型
			Str("ID", instID).           // 实例ID
			Str("DP", dataProvider).     // 数据源
			Int("Bar", barMinutes).      // K线周期（分钟）
			Str("UrlScheme", urlScheme). // UrlScheme
			Logger(),
	}
}

func NewLogger(typ string) zerolog.Logger {
	return l.log.With().
		Str("type", typ). // 日志类型
		Logger()
}

func (l *Logger) WithSubscribers(subscribers []config.Subscriber) *Logger {
	return &Logger{
		log: l.log.With().Interface("subscribers", subscribers).Logger(),
	}
}

// PrintError 打印错误日志。
func (l *Logger) PrintError(err error, disableNotify bool) {
	if errors.Is(err, context.Canceled) {
		l.PrintStrategyStop()
		return
	}
	l.log.Error().Err(err).
		Bool("disableNotify", disableNotify).Send()
}

func (l *Logger) PrintStrategyStop() {
	l.log.Warn().Msg("交易策略已停止")
}

// PrintStrategyMetrics 演示打印多个字段：时间、价格、RSI、MA5、MA20。
func (l *Logger) PrintStrategyMetrics(candleTime time.Time, price, curRSI, ma5, ma20 float64, abnormal string) {
	log := l.log.Info()
	if abnormal != "" {
		log = l.log.Warn().Str("异常", abnormal)
	}
	log.
		Time("Time", candleTime). // 时间
		Float64("Price", price).  // 价格
		Float64("RSI", curRSI).   // RSI
		Float64("MA5", ma5).      // MA5
		Float64("MA20", ma20).    // MA20
		Send()
}

// PrintErrorWithTime 打印包含特定时间的错误。(是否 Notify)
func (l *Logger) PrintErrorWithTime(candleTime time.Time, err error, disableNotify bool) {
	l.log.Error().
		Time("Time", candleTime). // 时间
		Err(err).
		Bool("disableNotify", disableNotify).
		Send()
}

// PrintBuyFail 记录买入失败。
func (l *Logger) PrintBuyFail(err error) {
	l.log.Error().Err(err).Msg("买入失败")
}

// PrintBuySuccess 记录买入成功并打印订单ID。
func (l *Logger) PrintBuySuccess(price, rsi float64, orderID string) {
	l.log.Warn().
		Float64("Price", price). // 价格
		Str("OrderID", orderID). // 订单ID
		Float64("RSI", rsi).     // RSI
		Msg("买入成功")
}

// PrintSellFail 记录卖出失败。
func (l *Logger) PrintSellFail(err error) {
	l.log.Error().Err(err).Msg("卖出失败")
}

// PrintSellSuccess 记录卖出成功并打印订单ID。
func (l *Logger) PrintSellSuccess(price, rsi float64, orderID string) {
	l.log.Warn().
		Float64("Price", price). // 价格
		Str("OrderID", orderID). // 订单ID
		Float64("RSI", rsi).     // RSI
		Msg("卖出成功")
}
