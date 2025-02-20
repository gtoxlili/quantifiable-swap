package trading

import "fmt"

var ErrInsufficientData = fmt.Errorf("采样数据不足")

var ErrFrequentSell = fmt.Errorf("卖出太频繁")

var ErrFrequentBuy = fmt.Errorf("买入太频繁")

var ErrInvalidSellCondition = fmt.Errorf("尚不满足卖出条件")

var ErrInvalidBuyCondition = fmt.Errorf("尚不满足买入条件")
