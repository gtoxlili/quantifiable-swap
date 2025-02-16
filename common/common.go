package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"math"
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

// ExtraPointsForInitialDecay 计算给定周期（period）下，使初始影响衰减至 0.1% 以下
// 所需的额外采样点数量
func ExtraPointsForInitialDecay(period int) int {
	// 目标衰减阈值（0.1%）
	threshold := 0.001
	base := float64(period-1) / float64(period)
	// 解方程 (base)^k <= threshold => k >= log(threshold) / log(base)
	k := math.Log(threshold) / math.Log(base)
	return int(math.Ceil(k)) + period
}
