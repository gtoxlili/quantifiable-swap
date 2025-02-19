package quantifiable

import (
	"context"
	"github.com/gtoxlili/quantifiable-swap/sequence"
)

// MA calculates a moving average over a specified window of candles.
type MA[T Number] struct {
	scale                int       // Number of candles per bar (e.g., if Bar is 15 and Frequency is 1, scale = 15)
	window               int       // Number of bars for the moving average window
	recentMAQueue        []float64 // Keeps a rolling history of the last few MA values
	sequence.Sequence[T]           // Embedded sequence providing candle data
	// 当前 MA 值
	currentMA float64
}

// NewMA initializes and returns a new MA instance.
func NewMA[T Number](window int, seq sequence.Sequence[T]) (Indicator[T], error) {
	ma := &MA[T]{
		window:        window,
		scale:         int(seq.Bar() / sequence.Frequency),
		Sequence:      seq,
		recentMAQueue: []float64{-1, -1, -1, -1, -1},
		currentMA:     -1,
	}
	return ma, nil
}

// CurrentVal returns the current moving average value.
func (ma *MA[T]) CurrentVal() float64 {
	return ma.currentMA
}

// PreviousVals returns a slice of previously calculated moving average values.
func (ma *MA[T]) PreviousVals() []float64 {
	return ma.recentMAQueue
}

// Update updates the embedded sequence and appends the newly calculated moving average to recentMAQueue.
func (ma *MA[T]) Update(ctx context.Context) (*sequence.Candle[T], error) {
	candle, err := ma.Sequence.Update(ctx)
	if err != nil {
		return nil, err
	}

	if candle.Time.Truncate(ma.Bar()).Equal(candle.Time) {
		// Append the most recent MA value
		ma.recentMAQueue = append(ma.recentMAQueue, ma.currentMA)
		if len(ma.recentMAQueue) > 5 {
			ma.recentMAQueue = ma.recentMAQueue[1:]
		}
		ma.currentMA = ma.calculateMA()
	}

	return candle, nil
}

// calculateMA computes the moving average by summing every "scale-th" candle over the specified window.
func (ma *MA[T]) calculateMA() float64 {
	candles := ma.Candles()
	required := ma.window * ma.scale
	if len(candles) < required {
		return -1
	}

	var sum float64
	for i := 0; i < required; i += ma.scale {
		sum += float64(candles[len(candles)-1-i].Value)
	}

	return sum / float64(ma.window)
}
