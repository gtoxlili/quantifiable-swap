package logger

import (
	"context"
	"errors"
	"github.com/gtoxlili/quantifiable-swap/bot"
	"github.com/gtoxlili/quantifiable-swap/common/logger/pretty/bark"
	"github.com/gtoxlili/quantifiable-swap/common/logger/pretty/console"
	"github.com/gtoxlili/quantifiable-swap/common/logger/pretty/tglog"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"github.com/rs/zerolog"
	"io"
	"os"
	"strconv"
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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.FloatingPointPrecision = 2

	var outSet []io.Writer
	outSet = append(outSet, console.NewConsoleWriter(os.Stdout, "15:04:05"))

	// todo: 待修改
	if bot.Bot != nil && constants.TGChatID != "" {
		chatId, _ := strconv.ParseInt(constants.TGChatID, 10, 64)
		outSet = append(outSet, tglog.NewBotWriter(bot.Bot, chatId))
	}

	if constants.BarkToken != "" {
		outSet = append(outSet, bark.NewNotifyWriter(constants.BarkToken))
	}

	l = &Logger{
		log: zerolog.New(zerolog.MultiLevelWriter(outSet...)).With().Timestamp().Logger(),
	}
}

// NewSwapLogger 创建一个新的 Logger 单例实例。
func NewSwapLogger(instID, dataProvider string, barMinutes int) *Logger {
	return &Logger{
		log: l.log.With().
			Str("type", "swap").     // 日志类型
			Str("ID", instID).       // 实例ID
			Str("DP", dataProvider). // 数据源
			Int("Bar", barMinutes).  // K线周期（分钟）
			Logger(),
	}
}

// NewGeneralLogger 打印代码逻辑中的日志
func NewGeneralLogger() zerolog.Logger {
	return l.log.With().
		Str("type", "general"). // 日志类型
		Logger()
}

func (l *Logger) WithSubscribers(subscribers []int64) *Logger {
	return &Logger{
		log: l.log.With().Ints64("subscribers", subscribers).Logger(),
	}
}

// PrintError 打印错误日志。
func (l *Logger) PrintError(err error, disableNotify bool) {
	if errors.Is(err, context.Canceled) {
		l.PrintWAPStop()
		return
	}
	l.log.Error().Err(err).
		Bool("disableNotify", disableNotify).Send()
}

func (l *Logger) PrintWAPStop() {
	l.log.Info().Msg("量化策略已停止")
}

// PrintIndicatorLog 演示打印多个字段：时间、价格、RSI、MA5、MA20。
func (l *Logger) PrintIndicatorLog(candleTime time.Time, price, curRSI, ma5, ma20 float64, abnormal string) {
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
