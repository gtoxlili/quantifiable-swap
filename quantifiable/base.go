package quantifiable

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

type IndicatorMetrics[T Number] interface {
	CurrentVal() float64
	PreviousVals() []float64
}

type Indicator[T Number] interface {
	sequence.Sequence[T]
	IndicatorMetrics[T]
}

type IndicatorBuilder[T Number] struct {
	seq     sequence.Sequence[T]
	metrics map[string]IndicatorMetrics[T]

	err error
}

func NewIndicatorBuilder[T Number](seq sequence.Sequence[T]) *IndicatorBuilder[T] {
	return &IndicatorBuilder[T]{
		seq:     seq,
		metrics: make(map[string]IndicatorMetrics[T]),
	}
}

func (b *IndicatorBuilder[T]) WithRSI(period int) *IndicatorBuilder[T] {
	if b.err != nil {
		return b
	}
	rsi, err := NewRSI(period, b.seq)
	if err != nil {
		b.err = fmt.Errorf("创建 RSI 包装器失败：%w", err)
	}
	b.seq = rsi
	b.metrics["RSI"] = rsi
	return b
}

func (b *IndicatorBuilder[T]) WithVol() *IndicatorBuilder[T] {
	if b.err != nil {
		return b
	}
	vol, err := NewVol(b.seq)
	if err != nil {
		b.err = fmt.Errorf("创建波动率包装器失败：%w", err)
	}
	b.seq = vol
	b.metrics["VOL"] = vol
	return b
}

type IndicatorDecorator[T Number] interface {
	Update() (*sequence.Candle[T], error)
	Indicator(name string) IndicatorMetrics[T]
}

func (b *IndicatorBuilder[T]) Build() (IndicatorDecorator[T], error) {
	if b.err != nil {
		return nil, b.err
	}
	return &IndicatorSet[T]{
		Sequence: b.seq,
		metrics:  b.metrics,
	}, nil
}

type IndicatorSet[T Number] struct {
	sequence.Sequence[T]
	metrics map[string]IndicatorMetrics[T]
}

func (s *IndicatorSet[T]) Indicator(name string) IndicatorMetrics[T] {
	return s.metrics[name]
}
