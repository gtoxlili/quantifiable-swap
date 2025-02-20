package handler

import (
	"bytes"
	"fmt"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/exchange"
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

func validateProviderInput(input string) (string, error) {
	prov := strings.TrimSpace(input)
	if prov == "" {
		return "", fmt.Errorf("❌ <b>输入错误</b>\n\n<i>提供商名称不能为空</i>")
	}

	availableProviders := exchange.ListAvailableProviders()
	for _, p := range availableProviders {
		if strings.EqualFold(p, prov) {
			return p, nil // 返回标准格式的提供商名称
		}
	}
	var providerList strings.Builder
	for _, p := range availableProviders {
		providerList.WriteString(fmt.Sprintf("• <code>%s</code>\n", p))
	}

	return "", fmt.Errorf("❌ <b>无效的提供商</b>\n\n支持的提供商：\n%s", providerList.String())
}

func formatJobPreview(job config.Job) string {
	var msgText string
	baseFormat := "🔑 ID: <code>%s</code>\n" +
		"📊 类型: <code>%s</code>\n" +
		"💱 交易对: <code>%s/%s</code>\n" +
		"%s" + // placeholder for type-specific info
		"⏱️ 采样间隔: <code>%s</code>"

	if job.Type == "monitor" {
		msgText = fmt.Sprintf(baseFormat,
			job.String(), job.Type, job.Symbol.Base, job.Symbol.Quote,
			fmt.Sprintf("📡 数据提供商: <code>%s</code>\n", job.Provider.Name),
			job.Bar,
		)
	} else {
		ordPb := job.Provider.InjectOrder
		if ordPb == "" {
			ordPb = job.Provider.Name
		}
		specificInfo := fmt.Sprintf(
			"💰 数量: 买入 <code>%.4f</code> / 卖出 <code>%.4f</code>\n"+
				"📡 数据提供商: <code>%s</code>\n"+
				"🏛️ 交易提供商: <code>%s</code>\n",
			job.Amount.Buy, job.Amount.Sell, job.Provider.Name, ordPb,
		)
		msgText = fmt.Sprintf(baseFormat,
			job.String(), job.Type, job.Symbol.Base, job.Symbol.Quote,
			specificInfo, job.Bar,
		)
	}
	return msgText
}

// 根据 Subscribers []int64 生成订阅者列表
func formatSubscribers(bot *tgApi.BotAPI, subs []int64) string {
	builder := &bytes.Buffer{}
	for _, sub := range subs {
		chat, err := bot.GetChat(tgApi.ChatInfoConfig{ChatConfig: tgApi.ChatConfig{ChatID: sub}})
		if err != nil || !chat.IsPrivate() {
			continue
		}
		builder.WriteString(fmt.Sprintf("@%s ", chat.UserName))
	}
	// 删除最后一个空格
	builder.Truncate(builder.Len() - 1)
	return builder.String()
}
