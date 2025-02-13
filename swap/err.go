package swap

import "fmt"

var ErrInsufficientSampleData = fmt.Errorf("采样数据不足")

var ErrSellTooFrequent = fmt.Errorf("卖出太频繁")

var ErrBuyTooFrequent = fmt.Errorf("买入太频繁")

var ErrNotMeetSellCondition = fmt.Errorf("尚不满足卖出条件")

var ErrNotMeetBuyCondition = fmt.Errorf("尚不满足买入条件")
