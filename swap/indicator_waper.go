package swap

import "context"

// IIndicatorWaper 接口
type IIndicatorWaper interface {
	Run(ctx context.Context)
	WithSubscribers(subscribers []int64)
	RunWithCustomPeriod(ctx context.Context, period int)
}
