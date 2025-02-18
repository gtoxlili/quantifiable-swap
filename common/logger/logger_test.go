package logger

import (
	"errors"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	// Test constants
	const (
		testID       = "BTC-USDT"
		testProvider = "okx"
		testBar      = 5
	)

	// Create logger instance
	logger := NewSwapLogger(testID, testProvider, testBar)
	// Test error logging
	testErr := errors.New("test error")
	logger.PrintError(testErr)

	// Test WAP stop
	logger.PrintWAPStop()

	// Test update price fail
	logger.PrintUpdatePriceFail(testErr)

	// Test indicator log
	testTime := time.Now()
	logger.PrintIndicatorLog(testTime, 50000.0, 90.5, 49000.0, 48000.0, "RSI")

	// Test error with time
	logger.PrintErrorWithTime(testTime, testErr)

	// Test buy operations
	logger.PrintBuyFail(testErr)
	logger.PrintBuySuccess(50000.0, 90.5, "order123")

	// Test sell operations
	logger.PrintSellFail(testErr)
	logger.PrintSellSuccess(50000.0, 90.5, "order123")

	time.Sleep(5 * time.Minute)
}
