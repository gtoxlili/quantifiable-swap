package tglog

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/common/logger/pretty"
	"io"
	"strings"
)

type BotWriter struct {
	bot     *tgApi.BotAPI
	chatID  int64
	entries chan pretty.LogData
}

func NewBotWriter(bot *tgApi.BotAPI, chatID int64) io.Writer {
	tw := &BotWriter{
		bot:     bot,
		chatID:  chatID,
		entries: make(chan pretty.LogData, 100), // 缓冲通道防止阻塞
	}
	go tw.processEntries() // 启动后台处理协程
	return tw
}

func (w *BotWriter) Write(p []byte) (n int, err error) {
	var logData pretty.LogData
	if err := json.Unmarshal(p, &logData); err != nil {
		return 0, fmt.Errorf("解析 Bot 日志失败: %v", err)
	}
	w.entries <- logData
	return len(p), nil
}

func (w *BotWriter) processEntries() {
	for logData := range w.entries {
		// 只打印交易日志
		if typ, ok := logData["type"].(string); !ok || typ != "swap" {
			continue
		}
		msgText := formatLogEntry(logData)

		msg := tgApi.NewMessage(w.chatID, msgText)
		msg.ParseMode = tgApi.ModeHTML
		// 置顶
		remote, err := w.bot.Send(msg)
		if err != nil {
			// 处理发送错误，可以添加重试逻辑
			fmt.Printf("发送 Bot 消息失败: %v\n", err)
			continue
		}
		if logData["level"] == "warn" || logData["level"] == "error" {
			w.alertPin(remote)
		}
	}
}

// 置顶警告
func (w *BotWriter) alertPin(remote tgApi.Message) {
	pinMsg := &tgApi.PinChatMessageConfig{
		ChatID:              w.chatID,
		MessageID:           remote.MessageID,
		DisableNotification: false,
	}
	_, err := w.bot.Request(pinMsg)
	if err != nil {
		fmt.Printf("置顶 Bot 消息失败: %v\n", err)
	}
}

func formatLogEntry(logData pretty.LogData) string {

	b := &bytes.Buffer{}

	// 基础信息块
	id := fmt.Sprintf("%v", logData["ID"])
	dp := fmt.Sprintf("%v", logData["DP"])
	bar := fmt.Sprintf("%v", logData["Bar"])
	timeVal := getTime(logData)

	// 标题部分
	b.WriteString(fmt.Sprintf("📊 <b>%s/%s</b>\n\n", dp, id))

	// 时间信息块
	b.WriteString(fmt.Sprintf("⏰ 系统时间: <code>%s</code>\n", timeVal))
	if t, ok := logData["Time"].(string); ok {
		b.WriteString(fmt.Sprintf("📅 采样时间: <code>%s</code>\n", t))
	}
	b.WriteString(fmt.Sprintf("⌛ Bar: <code>%sm</code>\n\n", bar))

	// 处理不同日志级别
	switch logData["level"] {
	case "error":
		handleError(b, logData)
	default:
		handleInfo(b, logData)
	}

	return b.String()
}

func handleError(b *bytes.Buffer, data pretty.LogData) {
	b.WriteString("🚨 <b>错误警报</b>\n\n")

	if msg, ok := data["message"].(string); ok {
		b.WriteString(fmt.Sprintf("📌 信息: <i>%s</i>\n", msg))
	}
	if errMsg, ok := data["error"].(string); ok {
		b.WriteString(fmt.Sprintf("💥 错误: <code>%s</code>\n", errMsg))
	}
	b.WriteString("\n")
}

func handleInfo(b *bytes.Buffer, data pretty.LogData) {
	if msg, ok := data["message"].(string); ok {
		switch {
		case strings.Contains(msg, "成功"):
			handleSuccess(b, data, msg)
		case strings.Contains(msg, "停止"):
			b.WriteString("⛔ <b>策略状态</b>\n")
			b.WriteString(fmt.Sprintf("📍 %s\n\n", msg))
		default:
			b.WriteString(fmt.Sprintf("ℹ️ %s\n\n", msg))
		}
	}

	// 如果有指标数据，添加空行后显示
	if hasMetrics(data) {
		handleMetrics(b, data)
	}
}

func handleSuccess(b *bytes.Buffer, data pretty.LogData, msg string) {
	var title string
	if strings.Contains(msg, "卖出") {
		title = "📤 <b>卖出订单</b>"
	} else if strings.Contains(msg, "买入") {
		title = "📥 <b>买入订单</b>"
	} else {
		title = "✅ <b>交易成功</b>"
	}

	b.WriteString(title + "\n")
	b.WriteString(fmt.Sprintf("📝 状态: %s\n", msg))

	if orderID, ok := data["OrderID"].(string); ok {
		b.WriteString(fmt.Sprintf("🎫 订单: <code>%s</code>\n", orderID))
	}
	if price, ok := data["Price"].(float64); ok {
		b.WriteString(fmt.Sprintf("💹 价格: <code>%.2f</code>\n", price))
	}
	b.WriteString("\n")
}

func handleMetrics(b *bytes.Buffer, data pretty.LogData) {
	b.WriteString("📊 <b>技术指标</b>\n")

	if price, ok := data["Price"].(float64); ok {
		b.WriteString(fmt.Sprintf("💰 当前价格: <code>%.2f</code>\n", price))
	}
	// 是否存在异常状态
	if abnormal, ok := data["异常"].(string); ok {
		// 可能是 RSI异常 或者 均线交叉异常
		b.WriteString(fmt.Sprintf("📢 指标信号: <code>%s</code>\n", abnormal))
	}

	// 合并展示移动平均线
	if ma5, ok := data["MA5"].(float64); ok {
		if ma20, ok := data["MA20"].(float64); ok {
			b.WriteString(fmt.Sprintf("📈 MA5/MA20: <code>%.2f</code> / <code>%.2f</code>\n", ma5, ma20))
		}
	}

	if rsi, ok := data["RSI"].(float64); ok {
		b.WriteString(fmt.Sprintf("🔋 RSI: <code>%.2f</code>\n", rsi))
	}
	b.WriteString("\n")
}

// 检查是否存在指标数据
func hasMetrics(data pretty.LogData) bool {
	keys := []string{"Price", "RSI", "MA5", "MA20"}
	for _, key := range keys {
		if _, ok := data[key].(float64); ok {
			return true
		}
	}
	return false
}

func getTime(data pretty.LogData) string {
	if t, ok := data["time"].(string); ok {
		return t
	}
	return "N/A"
}
