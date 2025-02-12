package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/client"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"net/http"
	"net/url"
)

func HmacSha256Sign(message, secrectKey string) ([]byte, error) {
	// 创建 HMAC-SHA256 哈希
	h := hmac.New(sha256.New, []byte(secrectKey))
	_, err := h.Write([]byte(message))
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC: %v", err)
	}

	return h.Sum(nil), nil
}

// Notify 发送 Bark 推送提醒
func Notify(title, message string, isCritical bool) error {
	baseURL := fmt.Sprintf(
		"https://api.day.app/%s/%s/%s",
		constants.BarkToken,
		url.QueryEscape(title),
		url.QueryEscape(message),
	)

	if isCritical {
		baseURL += "?level=critical&volume=5"
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
