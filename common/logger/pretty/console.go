package pretty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type ConsoleWriter struct {
	out        io.Writer
	timeFormat string
	entry      chan []byte
}

func NewConsoleWriter(out io.Writer, timeFormat string) *ConsoleWriter {
	cw := &ConsoleWriter{
		out:        out,
		timeFormat: timeFormat,
		entry:      make(chan []byte, 100),
	}
	go cw.processEntries()
	return cw
}

func (cw *ConsoleWriter) Write(p []byte) (n int, err error) {
	pCopy := make([]byte, len(p))
	copy(pCopy, p)
	cw.entry <- pCopy
	return len(p), nil
}

func (cw *ConsoleWriter) processEntries() {
	for entry := range cw.entry {
		msgBytes, err := formatLogEntry(entry)
		if err != nil {
			fmt.Printf("解析日志失败: %v\n", err)
			continue
		}

		cw.out.Write(msgBytes)
	}
}

func formatLogEntry(data []byte) ([]byte, error) {
	var logData map[string]interface{}
	if err := json.Unmarshal(data, &logData); err != nil {
		return nil, err
	}

	b := &bytes.Buffer{}

	handlePrefix(b, logData)
	divider(b)

	// 如果有时间
	if t, ok := logData["Time"].(string); ok {
		b.WriteString(fmt.Sprintf("Time: %s", colorize(t, colorYellow)))
		delete(logData, "Time")
		divider(b)
	}

	switch logData["level"] {
	case "error":
		handleError(b, logData)
	case "info":
		handleInfo(b, logData)
	}

	b.WriteString("\n")
	return b.Bytes(), nil
}

var (
	colorReset   = "\033[0m"
	colorBlue    = "\033[1;34m"
	colorRed     = "\033[1;31m"
	colorMagenta = "\033[1;35m"
	colorCyan    = "\033[1;36m"
	colorGreen   = "\033[1;32m"
	colorYellow  = "\033[1;33m"
)

func colorize(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, colorReset)
}

func formatSegment(text, color string) string {
	return fmt.Sprintf("[%s]", colorize(text, color))
}

func handlePrefix(b *bytes.Buffer, data map[string]interface{}) {
	if instID, ok := data["ID"]; ok {
		b.WriteString(formatSegment(instID.(string), colorCyan))
		delete(data, "ID")
	}
	if dp, ok := data["DP"]; ok {
		b.WriteString(formatSegment(dp.(string), colorMagenta))
		delete(data, "DP")
	}
	if bar, ok := data["Bar"].(float64); ok {
		var barColor string
		switch bar {
		case 15:
			barColor = colorGreen
		case 60:
			barColor = colorYellow
		default:
			barColor = colorBlue
		}
		b.WriteString(formatSegment(fmt.Sprintf("%dm", int(bar)), barColor))
		delete(data, "Bar")
	}
}

// 分割符
func divider(b *bytes.Buffer) {
	b.WriteString(colorize(" | ", colorBlue))
}

func handleError(b *bytes.Buffer, data map[string]interface{}) {
	title := "Error"
	if msg, ok := data["message"].(string); ok {
		title = msg
		delete(data, "message")
	}
	title += ": "
	b.WriteString(title)
	if errMsg, ok := data["error"].(string); ok {
		b.WriteString(fmt.Sprintf("%s", colorize(errMsg, colorRed)))
		delete(data, "error")
		divider(b)
	}
	handleExtraFields(b, data)
	// 删除 最后一个 divider(b)
	b.Truncate(b.Len() - len(colorize(" | ", colorBlue)))
}

func handleInfo(b *bytes.Buffer, data map[string]interface{}) {
	if msg, ok := data["message"].(string); ok {
		b.WriteString(colorize(msg, colorGreen))
		delete(data, "message")
		divider(b)
	}
	handleExtraFields(b, data)
	// 删除 最后一个 divider(b)
	b.Truncate(b.Len() - len(colorize(" | ", colorBlue)))
}

func hashColor(s string) string {
	var sum int
	for _, c := range s {
		sum += int(c) // 累加每个字符的 Unicode 码
	}
	return fmt.Sprintf("\033[1;%dm", 31+sum%6)
}

// 打印多余的字段
func handleExtraFields(b *bytes.Buffer, data map[string]interface{}) {
	for k, v := range data {
		if k == "level" || k == "time" {
			continue
		}
		var vv string
		switch v.(type) {
		case string:
			vv = v.(string)
		case float64:
			vv = fmt.Sprintf("%.2f", v)
		default:
			vv = fmt.Sprintf("%v", v)
		}
		b.WriteString(fmt.Sprintf("%s: %s", k, colorize(vv, hashColor(vv))))
		divider(b)
	}
}
