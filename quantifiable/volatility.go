package quantifiable

import (
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"math"
)

type Vol[T Number] struct {
	scale int
	sequence.Sequence[T]
}

func NewVol[T Number](seq sequence.Sequence[T]) (Indicator[T], error) {
	vol := &Vol[T]{
		scale:    int(seq.Bar() / sequence.Frequency),
		Sequence: seq,
	}
	return vol, nil
}

func (vol *Vol[T]) CurrentVal() float64 {
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

func (vol *Vol[T]) PreviousVals() []float64 {
	panic("implement me")
}

func (vol *Vol[T]) Update() (*sequence.Candle[T], error) {
	// defer fmt.Println("Vol Update")
	return vol.Sequence.Update()
}
