package sequence

import (
	"context"
	"time"
)

type Candle[T any] struct {
	Time  time.Time
	Value T
}

// Sequence 序列协议
type Sequence[T any] interface {
	// Candles 返回序列
	Candles() []Candle[T]
	// Update 更新序列
	Update(ctx context.Context) (*Candle[T], error)
	// Bar 获取 Bar
	Bar() time.Duration
	LastBarIndex() int
}
