package quantifiable

import (
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/exp/constraints"
	"math"
)

type VolIndicator[T constraints.Integer | constraints.Float] interface {
	sequence.Sequence[T]
	CurrentVol() float64
}

type Vol[T constraints.Integer | constraints.Float] struct {
	scale int
	sequence.Sequence[T]
}

func NewVol[T constraints.Integer | constraints.Float](seq sequence.Sequence[T]) (VolIndicator[T], error) {
	vol := &Vol[T]{
		scale:    int(seq.Bar() / sequence.Frequency),
		Sequence: seq,
	}
	return vol, nil
}

func (vol *Vol[T]) CurrentVol() float64 {
	return vol.calculateVol()
}

func (vol *Vol[T]) calculateVol() float64 {
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
