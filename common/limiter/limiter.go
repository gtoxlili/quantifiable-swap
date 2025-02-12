package limiter

import (
	"context"
	"golang.org/x/time/rate"
	"time"
)

// RateLimiter 是一个简单的限速器接口
type RateLimiter interface{ Wait() error }

// tokenRateLimiter 使用令牌桶算法
type tokenRateLimiter struct{ limiter *rate.Limiter }

// NewTokenRateLimiter 创建一个令牌桶算法的 RateLimiter。
// 例如：限速 10 次/秒，突发（burst）也允许一次性消耗到 10 个令牌
func NewTokenRateLimiter(rps int) RateLimiter {
	return &tokenRateLimiter{limiter: rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), 1)}
}

func NewTokenRateLimiterWithBurst(rps, burst int) RateLimiter {
	return &tokenRateLimiter{limiter: rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), burst)}
}

func (t tokenRateLimiter) Wait() error {
	// 在这里可根据需要自定义 context，例如设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 每次请求需要消耗 1 个令牌
	// 如果没有可用令牌就会阻塞直到超时或拿到令牌为止
	return t.limiter.Wait(ctx)
}
