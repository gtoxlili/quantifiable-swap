package indicator

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/sequence"
)

type Builder[T Number] struct {
	seq     sequence.Sequence[T]
	metrics map[string]Metrics[T]

	err error
}

func NewIndicatorBuilder[T Number](seq sequence.Sequence[T]) *Builder[T] {
	return &Builder[T]{
		seq:     seq,
		metrics: make(map[string]Metrics[T]),
	}
}

func (b *Builder[T]) WithRSI(period int) *Builder[T] {
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

func (b *Builder[T]) WithMA(window int) *Builder[T] {
	if b.err != nil {
		return b
	}
	ma, err := NewMA(window, b.seq)
	if err != nil {
		b.err = fmt.Errorf("创建 MA%d 包装器失败：%w", window, err)
	}
	b.seq = ma
	b.metrics[fmt.Sprintf("MA%d", window)] = ma
	return b
}

func (b *Builder[T]) WithCustom(name string, fn func(seq sequence.Sequence[T]) (Indicator[T], error)) *Builder[T] {
	if b.err != nil {
		return b
	}
	ind, err := fn(b.seq)
	if err != nil {
		b.err = fmt.Errorf("创建[%s]包装器失败：%w", name, err)
	}
	b.seq = ind
	b.metrics[name] = ind
	return b
}

func (b *Builder[T]) Build() (Decorator[T], error) {
	if b.err != nil {
		return nil, b.err
	}
	return &Set[T]{
		Sequence: b.seq,
		metrics:  b.metrics,
	}, nil
}

type Set[T Number] struct {
	sequence.Sequence[T]
	metrics map[string]Metrics[T]
}

func (s *Set[T]) Indicator(name string) (Metrics[T], error) {
	if m, ok := s.metrics[name]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("未找到指标 [%s]", name)
}
