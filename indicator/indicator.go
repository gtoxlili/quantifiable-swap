package indicator

import (
	"context"
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

type Metrics[T Number] interface {
	CurrentVal() float64
	PreviousVals() []float64
}

type Indicator[T Number] interface {
	sequence.Sequence[T]
	Metrics[T]
}

type Decorator[T Number] interface {
	Update(ctx context.Context) (*sequence.Candle[T], error)
	Indicator(name string) (Metrics[T], error)
}
