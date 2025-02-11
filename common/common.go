package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"quantifiable-swap/client"
	"quantifiable-swap/constants"
)

func HmacSha256Sign(message, secrectKey string) (string, error) {
	// 创建 HMAC-SHA256 哈希
	h := hmac.New(sha256.New, []byte(secrectKey))
	_, err := h.Write([]byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to generate HMAC: %v", err)
	}

	// 计算 HMAC 并进行 Base64 编码
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
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
