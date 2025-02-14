package quantifiable

import (
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/exp/constraints"
	"math"
)

type VolatilityIndicator[T constraints.Integer | constraints.Float] interface {
	sequence.Sequence[T]
	CurrentVolatility() float64
}

type Volatility[T constraints.Integer | constraints.Float] struct {
	scale int
	sequence.Sequence[T]
}

func NewVolatility[T constraints.Integer | constraints.Float](seq sequence.Sequence[T]) (VolatilityIndicator[T], error) {
	vol := &Volatility[T]{
		scale:    int(seq.Bar() / sequence.Frequency),
		Sequence: seq,
	}
	return vol, nil
}

func (vol *Volatility[T]) CurrentVolatility() float64 {
	return vol.calculateVolatility()
}

func (vol *Volatility[T]) calculateVolatility() float64 {
	candles := vol.Candles()
	if len(candles) < vol.scale || vol.scale <= 1 {
		return 0
	}
	aggregatedData := candles[len(candles)-vol.scale:]

	var sum float64
	for _, c := range aggregatedData {
		sum += float64(c.Value)
	}
	mean := sum / float64(len(aggregatedData))

	var varianceSum float64
	for _, c := range aggregatedData {
		diff := float64(c.Value) - mean
		varianceSum += diff * diff
	}
	// 注意：样本标准差一般用 (n-1)，如果希望总体标准差，可改为 n
	variance := varianceSum / float64(len(aggregatedData)-1)

	return math.Sqrt(variance)
}
