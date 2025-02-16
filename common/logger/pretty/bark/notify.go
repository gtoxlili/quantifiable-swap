package bark

import (
	"encoding/json"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/client"
	"net/http"
	"net/url"
)

type BarkWriter struct {
	token   string
	entries chan []byte
}

func NewBarkWriter(token string) *BarkWriter {
	bw := &BarkWriter{
		token:   token,
		entries: make(chan []byte, 100),
	}
	go bw.processEntries()
	return bw
}

func (w *BarkWriter) Write(p []byte) (n int, err error) {
	pCopy := make([]byte, len(p))
	copy(pCopy, p)
	w.entries <- pCopy
	return len(p), nil
}

func (w *BarkWriter) processEntries() {
	for entry := range w.entries {
		var logData map[string]interface{}
		if err := json.Unmarshal(entry, &logData); err != nil {
			fmt.Printf("解析 Bark 日志失败: %v\n", err)
			continue
		}
		if logData["level"] == "warn" || logData["level"] == "error" {
			title, message, groupName := formatLogEntry(logData)
			if err := w.notify(title, message, groupName); err != nil {
				fmt.Printf("发送 Bark 消息失败: %v\n", err)
			}
		}
	}
}

func formatLogEntry(logData map[string]interface{}) (title, message, groupName string) {
	switch logData["level"] {
	case "warn":
		return handleWarn(logData)
	case "error":
		return handleError(logData)
	default:
		panic("unexpected log level")
	}
}

func handleError(logData map[string]interface{}) (title, message, groupName string) {
	id := fmt.Sprintf("%v", logData["ID"])
	err := fmt.Sprintf("%v", logData["error"])

	if msg, ok := logData["message"]; ok {
		return fmt.Sprintf("[%s] %s", id, msg), fmt.Sprintf("Error: %s", err), "自动交易"
	}

	return fmt.Sprintf("[%s] %s", id, err), "", "自动交易"
}

func handleWarn(logData map[string]interface{}) (title, message, groupName string) {
	id := fmt.Sprintf("%v", logData["ID"])

	if msg, ok := logData["message"]; ok {
		return fmt.Sprintf("[%s] %s", id, msg), fmt.Sprintf("Price: %.2f, RSI: %.2f, OrderId: %s", logData["Price"], logData["RSI"], logData["OrderID"]), "自动交易"
	}

	abnormal := logData["异常"].(string)
	if abnormal == "RSI" {
		return fmt.Sprintf("[%s] RSI提醒", id),
			fmt.Sprintf("[%s][%s] Price: %.2f, RSI: %.2f",
				logData["DP"],
				fmt.Sprintf("%dm", logData["Bar"]),
				logData["Price"], logData["RSI"]),
			"指标监控"
	} else if abnormal == "金叉" || abnormal == "死叉" {
		return fmt.Sprintf("[%s] 均线交叉提醒", id),
			fmt.Sprintf("[%s][%s]%s Price: %.2f, MA5: %.2f, MA20: %.2f",
				logData["DP"],
				fmt.Sprintf("%dm", logData["Bar"]), abnormal,
				logData["Price"], logData["MA5"], logData["MA20"],
			),
			"指标监控"
	}

	panic("unexpected abnormal")
}

func (w *BarkWriter) notify(title, message string, groupName string) error {
	baseURL := fmt.Sprintf(
		"https://api.day.app/%s/%s",
		w.token,
		url.QueryEscape(title),
	)
	if message != "" {
		baseURL = baseURL + "/" + url.QueryEscape(message)
	}
	if groupName != "" {
		baseURL = baseURL + "?group=" + url.QueryEscape(groupName)
	}

	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return fmt.Errorf("create notify request failed: %w", err)
	}

	resp, err := client.C.Do(req)
	if err != nil {
		return fmt.Errorf("notify request failed: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
