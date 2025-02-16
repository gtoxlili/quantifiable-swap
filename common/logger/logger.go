package logger

import (
	"fmt"
	"time"
)

// ANSI color codes
var (
	colorReset   = "\033[0m"
	colorBlue    = "\033[1;34m"
	colorRed     = "\033[1;31m"
	colorMagenta = "\033[1;35m"
	colorCyan    = "\033[1;36m"
	colorGreen   = "\033[1;32m"
	colorYellow  = "\033[1;33m"
	colorWhite   = "\033[1;37m"
	colorBold    = "\033[1m"
)

// Logger is a logger that encapsulates colorful log printing logic.
type Logger struct {
	InstID       string
	DataProvider string
	BarMinutes   int
}

// NewLogger creates a new Logger.
func NewLogger(instID, dataProvider string, barMinutes int) *Logger {
	return &Logger{
		InstID:       instID,
		DataProvider: dataProvider,
		BarMinutes:   barMinutes,
	}
}

func bracket(text, textColor string) string {
	// Bold brackets, colored text, then reset all styles at the end.
	return fmt.Sprintf(
		"%s[%s%s%s%s]%s",
		colorBold, textColor, text, colorReset, colorBold, colorReset,
	)
}

// buildPrefix returns a formatted prefix string with different colors
// for InstID, DataProvider, and barMinutes (depending on their intervals).
func (l *Logger) buildPrefix() string {
	var barColor string
	switch l.BarMinutes {
	case 15:
		barColor = colorGreen
	case 60:
		barColor = colorYellow
	case 240:
		barColor = colorBlue
	default:
		barColor = colorWhite
	}
	instPart := bracket(l.InstID, colorCyan)
	dpPart := bracket(l.DataProvider, colorMagenta)
	barPart := bracket(fmt.Sprintf("%dm", l.BarMinutes), barColor)
	return instPart + dpPart + barPart
}

// PrintError prints an error log with our new prefix.
func (l *Logger) PrintError(err error) {
	fmt.Printf("%s %s|%s %sError:%s %s%v%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorBold, colorReset,
		colorRed, err, colorReset)
}

// PrintRSIWAPStop indicates that the RSIWAP strategy has stopped.
func (l *Logger) PrintRSIWAPStop() {
	fmt.Printf("%s %s|%s %sRSIWAP 策略已停止%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorMagenta, colorReset)
}

// PrintUpdatePriceFail logs a failure to update price series.
func (l *Logger) PrintUpdatePriceFail(err error) {
	fmt.Printf("%s %s|%s 更新价格序列失败:%s %s%v%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorBold, colorRed, err, colorReset)
}

// PrintIndicatorLog demonstrates printing multiple fields: time, price, RSI, MA5, MA20.
func (l *Logger) PrintIndicatorLog(candleTime time.Time, price, curRSI, ma5, ma20 float64) {
	fmt.Printf("%s %s|%s Time: %s%s%s %s|%s Price: %s%.2f%s %s|%s RSI: %s%.2f%s %s|%s MA5: %s%.2f%s %s|%s MA20: %s%.2f%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorYellow, candleTime.Format("15:04:05"), colorReset,
		colorBlue, colorReset,
		colorGreen, price, colorReset,
		colorBlue, colorReset,
		colorRed, curRSI, colorReset,
		colorBlue, colorReset,
		colorMagenta, ma5, colorReset,
		colorBlue, colorReset,
		colorCyan, ma20, colorReset)
}

// PrintErrorWithTime prints an error with a specific time included.
func (l *Logger) PrintErrorWithTime(candleTime time.Time, err error) {
	fmt.Printf("%s %s|%s Time: %s%s%s %s|%s Error:%s %s%v%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorYellow, candleTime.Format("15:04:05"), colorReset,
		colorBlue, colorReset,
		colorBold, colorReset,
		colorRed, err, colorReset)
}

// PrintBuyFail logs a buy failure.
func (l *Logger) PrintBuyFail(err error) {
	fmt.Printf("%s %s|%s 买入失败:%s %s%v%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorBold, colorRed, err, colorReset)
}

// PrintBuySuccess logs a buy success and prints the order ID.
func (l *Logger) PrintBuySuccess(orderID string) {
	fmt.Printf("%s %s|%s %s买入成功%s %s|%s 订单号: %s%s%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorGreen, colorReset,
		colorBlue, colorReset,
		colorCyan, orderID, colorReset)
}

// PrintSellFail logs a sell failure.
func (l *Logger) PrintSellFail(err error) {
	fmt.Printf("%s %s|%s 卖出失败:%s %s%v%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorBold, colorRed, err, colorReset)
}

// PrintSellSuccess logs a sell success and prints the order ID.
func (l *Logger) PrintSellSuccess(orderID string) {
	fmt.Printf("%s %s|%s %s卖出成功%s %s|%s 订单号: %s%s%s\n",
		l.buildPrefix(),
		colorBlue, colorReset,
		colorGreen, colorReset,
		colorBlue, colorReset,
		colorCyan, orderID, colorReset)
}
