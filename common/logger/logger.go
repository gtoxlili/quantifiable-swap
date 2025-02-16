package logger

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

// Logger is a logger that encapsulates logging logic using zerolog.
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

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		NoColor:    false,
	}

	l = &Logger{
		log: zerolog.New(output).With().Timestamp().Logger(),
	}
}

// NewLogger creates a new Logger singleton instance.
func NewLogger(instID, dataProvider string, barMinutes int) *Logger {
	return &Logger{
		log: l.log.With().
			Str("INST", instID).
			Str("DP", dataProvider).
			Int("BAR", barMinutes).
			Logger(),
	}
}

// PrintError prints an error log.
func (l *Logger) PrintError(err error) {
	l.log.Error().Err(err).Send()
}

func (l *Logger) PrintWAPStop() {
	l.log.Info().Msg("量化策略已停止")
}

// PrintUpdatePriceFail logs a failure to update price series.
func (l *Logger) PrintUpdatePriceFail(err error) {
	l.log.Error().Err(err).Msg("更新价格序列失败")
}

// PrintIndicatorLog demonstrates printing multiple fields: time, price, RSI, MA5, MA20.
func (l *Logger) PrintIndicatorLog(candleTime time.Time, price, curRSI, ma5, ma20 float64) {
	l.log.Info().
		Time("T", candleTime).
		Float64("P", price).
		Float64("RSI", curRSI).
		Float64("MA5", ma5).
		Float64("MA20", ma20).
		Send()
}

// PrintErrorWithTime prints an error with a specific time included.
func (l *Logger) PrintErrorWithTime(candleTime time.Time, err error) {
	l.log.Error().
		Time("T", candleTime).
		Err(err).
		Send()
}

// PrintBuyFail logs a buy failure.
func (l *Logger) PrintBuyFail(err error) {
	l.log.Error().Err(err).Msg("买入失败")
}

// PrintBuySuccess logs a buy success and prints the order ID.
func (l *Logger) PrintBuySuccess(orderID string) {
	l.log.Info().
		Str("OID", orderID).
		Msg("买入成功")
}

// PrintSellFail logs a sell failure.
func (l *Logger) PrintSellFail(err error) {
	l.log.Error().Err(err).Msg("卖出失败")
}

// PrintSellSuccess logs a sell success and prints the order ID.
func (l *Logger) PrintSellSuccess(orderID string) {
	l.log.Info().
		Str("OID", orderID).
		Msg("卖出成功")
}
