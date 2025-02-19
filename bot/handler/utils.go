package handler

import (
	"fmt"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
)

// sendMessage is a convenience method for sending plain text messages.
func (handler *BotHandler) sendMessage(chatID int64, text string) {
	msg := tgApi.NewMessage(chatID, text)
	handler.BotAPI.Send(msg)
}

// validateSymbolInput checks if input is in "BASE/QUOTE" format and returns the two parts.
func validateSymbolInput(input string) (string, string, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("格式错误，请使用 BASE/QUOTE 格式")
	}
	base := strings.TrimSpace(parts[0])
	quote := strings.TrimSpace(parts[1])
	// Add more validation if necessary.
	return base, quote, nil
}

// validateAmountInput checks if input is in "buy/sell" float format and returns both floats.
func validateAmountInput(input string) (float64, float64, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("格式错误，请使用 买入数量/卖出数量 格式")
	}
	buyAmount, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("买入数量格式错误: %v", err)
	}
	sellAmount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("卖出数量格式错误: %v", err)
	}
	return buyAmount, sellAmount, nil
}
