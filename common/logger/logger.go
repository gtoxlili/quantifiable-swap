package logger

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/bot"
	"github.com/gtoxlili/quantifiable-swap/common"
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
	instID string
	dp     string
	bar    int
	log    zerolog.Logger
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

	if bot.Bot != nil && constants.TGChatID != "" {
		chatId, _ := strconv.ParseInt(constants.TGChatID, 10, 64)
		outSet = append(outSet, tglog.NewTelegramWriter(bot.Bot, chatId))
	}

	l = &Logger{
		log: zerolog.New(zerolog.MultiLevelWriter(outSet...)).With().Timestamp().Logger(),
	}
}

// NewLogger 创建一个新的 Logger 单例实例。
func NewLogger(instID, dataProvider string, barMinutes int) *Logger {
	return &Logger{
		instID: instID,
		dp:     dataProvider,
		bar:    barMinutes,
		log: l.log.With().
			Str("ID", instID).       // 实例ID
			Str("DP", dataProvider). // 数据源
			Int("Bar", barMinutes).  // K线周期（分钟）
			Logger(),
	}
}

// PrintError 打印错误日志。
func (l *Logger) PrintError(err error) {
	l.log.Error().Err(err).Send()
}

func (l *Logger) PrintWAPStop() {
	l.log.Info().Msg("量化策略已停止")
}

// PrintUpdatePriceFail 记录更新价格序列失败。
func (l *Logger) PrintUpdatePriceFail(err error) {
	l.log.Error().Err(err).Msg("更新价格序列失败")
}

// PrintIndicatorLog 演示打印多个字段：时间、价格、RSI、MA5、MA20。
func (l *Logger) PrintIndicatorLog(candleTime time.Time, price, curRSI, ma5, ma20 float64, alert string) {
	log := l.log.Info()
	if alert != "" {
		log = l.log.Warn()
	}
	log.
		Time("Time", candleTime). // 时间
		Float64("Price", price).  // 价格
		Float64("RSI", curRSI).   // RSI
		Float64("MA5", ma5).      // MA5
		Float64("MA20", ma20).    // MA20
		Send()
	if alert == "RSI" {
		common.Notify(
			"⚠️ ["+l.instID+"] RSI提醒",
			fmt.Sprintf("[%s][%s] Price: %.2f, RSI: %.2f",
				l.dp,
				fmt.Sprintf("%dm", l.bar),
				price, curRSI),
			"指标监控",
		)
	} else {
		common.Notify(
			"⚠️ ["+l.instID+"] 均线交叉提醒",
			fmt.Sprintf("[%s][%s]%s Price: %.2f, MA5: %.2f, MA20: %.2f",
				l.dp,
				fmt.Sprintf("%dm", l.bar), alert,
				price, ma5, ma20),
			"指标监控",
		)
	}
}

// PrintErrorWithTime 打印包含特定时间的错误。
func (l *Logger) PrintErrorWithTime(candleTime time.Time, err error) {
	l.log.Error().
		Time("Time", candleTime). // 时间
		Err(err).
		Send()
}

// PrintBuyFail 记录买入失败。
func (l *Logger) PrintBuyFail(err error) {
	l.log.Error().Err(err).Msg("买入失败")
	common.Notify(
		"❌ ["+l.instID+"]自动买入失败",
		fmt.Sprintf("Error: %v", err),
		"自动交易",
	)
}

// PrintBuySuccess 记录买入成功并打印订单ID。
func (l *Logger) PrintBuySuccess(price, rsi float64, orderID string) {
	l.log.Info().
		Float64("Price", price). // 价格
		Str("OrderID", orderID). // 订单ID
		Float64("RSI", rsi).     // RSI
		Msg("买入成功")
	common.Notify(
		"🚀 ["+l.instID+"]自动买入提醒",
		fmt.Sprintf("Price: %.2f, RSI: %.2f, OrderId: %s", price, rsi, orderID),
		"自动交易",
	)
}

// PrintSellFail 记录卖出失败。
func (l *Logger) PrintSellFail(err error) {
	l.log.Error().Err(err).Msg("卖出失败")
	common.Notify(
		"❌ ["+l.instID+"]自动卖出失败",
		fmt.Sprintf("Error: %v", err),
		"自动交易",
	)
}

// PrintSellSuccess 记录卖出成功并打印订单ID。
func (l *Logger) PrintSellSuccess(price, rsi float64, orderID string) {
	l.log.Info().
		Float64("Price", price). // 价格
		Str("OrderID", orderID). // 订单ID
		Float64("RSI", rsi).     // RSI
		Msg("卖出成功")
	common.Notify(
		"🚀 ["+l.instID+"]自动卖出提醒",
		fmt.Sprintf("Price: %.2f, RSI: %.2f, OrderId: %s", price, rsi, orderID),
		"自动交易",
	)
}
