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
	msg.ParseMode = tgApi.ModeHTML
	handler.BotAPI.Send(msg)
}

func (handler *BotHandler) sendEditMessage(chatID int64, messageID int, text string) {
	msg := tgApi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = tgApi.ModeHTML
	handler.BotAPI.Send(msg)
}

// 带 ReplyMarkup
func (handler *BotHandler) sendMessageWithMarkup(chatID int64, text string, markup tgApi.InlineKeyboardMarkup) {
	msg := tgApi.NewMessage(chatID, text)
	msg.ParseMode = tgApi.ModeHTML
	msg.ReplyMarkup = markup
	handler.BotAPI.Send(msg)
}

func validateSymbolInput(input string) (string, string, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("❌ <b>格式错误</b>\n\n请使用格式：<code>BASE/QUOTE</code>")
	}
	base := strings.TrimSpace(parts[0])
	quote := strings.TrimSpace(parts[1])
	if base == "" || quote == "" {
		return "", "", fmt.Errorf("❌ <b>格式错误</b>\n\n<i>交易对不能为空</i>")
	}
	return base, quote, nil
}

func validateAmountInput(input string) (float64, float64, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("❌ <b>格式错误</b>\n\n请使用格式：<code>买入数量/卖出数量</code>")
	}
	buyAmount, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("❌ <b>数值错误</b>\n\n买入数量：<i>%v</i>", err)
	}
	sellAmount, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("❌ <b>数值错误</b>\n\n卖出数量：<i>%v</i>", err)
	}
	if buyAmount <= 0 || sellAmount <= 0 {
		return 0, 0, fmt.Errorf("❌ <b>数值错误</b>\n\n<i>交易数量必须大于0</i>")
	}
	return buyAmount, sellAmount, nil
}
