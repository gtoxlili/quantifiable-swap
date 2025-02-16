package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

type TelegramWriter struct {
	bot     *tgbotapi.BotAPI
	chatID  int64
	entries chan []byte
}

func NewTelegramWriter(bot *Bot, chatID int64) *TelegramWriter {
	tw := &TelegramWriter{
		bot:     bot.BotAPI,
		chatID:  chatID,
		entries: make(chan []byte, 100), // 缓冲通道防止阻塞
	}
	go tw.processEntries() // 启动后台处理协程
	return tw
}

func (w *TelegramWriter) Write(p []byte) (n int, err error) {
	pCopy := make([]byte, len(p))
	copy(pCopy, p)
	w.entries <- pCopy
	return len(p), nil
}

func (w *TelegramWriter) processEntries() {
	for entry := range w.entries {
		msgText, err := formatLogEntry(entry)
		if err != nil {
			fmt.Printf("解析日志失败: %v\n", err)
			continue
		}

		msg := tgbotapi.NewMessage(w.chatID, msgText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		_, err = w.bot.Send(msg)
		if err != nil {
			// 处理发送错误，可以添加重试逻辑
			fmt.Printf("发送消息失败: %v\n", err)
		}
	}
}

func formatLogEntry(data []byte) (string, error) {
	var logData map[string]interface{}
	if err := json.Unmarshal(data, &logData); err != nil {
		return "", err
	}

	b := &bytes.Buffer{}

	// 公共字段
	id := fmt.Sprintf("%v", logData["ID"])
	dp := fmt.Sprintf("%v", logData["DP"])
	bar := fmt.Sprintf("%v", logData["Bar"])
	timeVal := getTime(logData)

	b.WriteString(fmt.Sprintf("🕒 *%s*\n", timeVal))
	b.WriteString(fmt.Sprintf("🔗 交易对: `%s`\n", id))
	b.WriteString(fmt.Sprintf("🏷 平台: %s | Bar: %s\n", dp, bar))

	// 交易时间
	if t, ok := logData["Time"].(string); ok {
		b.WriteString(fmt.Sprintf("🕒 采样时间: `%s`\n", t))
	}

	// 处理不同日志级别
	switch logData["level"] {
	case "error":
		b.WriteString("❗️ *ERROR*\n")
		if msg, ok := logData["message"].(string); ok {
			b.WriteString(fmt.Sprintf("📝 _%s_\n", msg))
		}
		if errMsg, ok := logData["error"].(string); ok {
			b.WriteString(fmt.Sprintf("🔥 ERROR: `%s`\n", errMsg))
		}

	case "info":
		if msg, ok := logData["message"].(string); ok {
			switch {
			case strings.Contains(msg, "成功"):
				handleSuccess(b, logData, msg)
			case strings.Contains(msg, "停止"):
				b.WriteString("🛑 策略状态:\n")
				b.WriteString(fmt.Sprintf("▫️ %s\n", msg))
			default:
				b.WriteString(fmt.Sprintf("ℹ️ %s\n", msg))
			}
		}
		handleMetrics(b, logData)
	}

	return b.String(), nil
}

func handleSuccess(b *bytes.Buffer, data map[string]interface{}, msg string) {
	emoji := "✅"
	if strings.Contains(msg, "卖出") {
		emoji = "📤"
	} else if strings.Contains(msg, "买入") {
		emoji = "📥"
	}

	b.WriteString(fmt.Sprintf("%s %s\n", emoji, msg))
	if orderID, ok := data["OrderID"].(string); ok {
		b.WriteString(fmt.Sprintf("🔖 订单号: `%s`\n", orderID))
	}
	if price, ok := data["Price"].(float64); ok {
		b.WriteString(fmt.Sprintf("💰 价格: `%.2f`\n", price))
	}
}

func handleMetrics(b *bytes.Buffer, data map[string]interface{}) {
	if rsi, ok := data["RSI"].(float64); ok {
		b.WriteString(fmt.Sprintf("📈 RSI: `%.2f` | ", rsi))
	}
	if ma5, ok := data["MA5"].(float64); ok {
		b.WriteString(fmt.Sprintf("MA5: `%.2f` | ", ma5))
	}
	if ma20, ok := data["MA20"].(float64); ok {
		b.WriteString(fmt.Sprintf("MA20: `%.2f` | ", ma20))
	}
	// 删除最后一个分隔符
	if b.Len() > 0 {
		b.Truncate(b.Len() - 3)
		b.WriteString("\n")
	}
}

func getTime(data map[string]interface{}) string {
	if t, ok := data["time"].(string); ok {
		return t
	}
	return "N/A"
}
