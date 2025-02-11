package quantifiable

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"quantifiable-swap/sequence"
)

// avgGain / avgLoss
type gainLoss struct {
	avgGain, avgLoss float64
}

// RSI Pack 对于序列的 RSI HOOK
type RSI[T constraints.Integer | constraints.Float] struct {
	period int
	sequence.Sequence[T]
	historyAvg []gainLoss
	// 存储最近 5 个 rsi 用于数据分析（队列）
	lastRSIQueue []float64
}

// NewRSIHOOK NewRSI 创建一个 RSI 包装器
func NewRSIHOOK[T constraints.Integer | constraints.Float](period int, seq sequence.Sequence[T]) (*RSI[T], error) {
	scale := int(seq.Bar() / sequence.Frequency)
	rsi := &RSI[T]{
		period:       period,
		Sequence:     seq,
		historyAvg:   make([]gainLoss, scale),
		lastRSIQueue: []float64{-1, -1, -1, -1, -1},
	}
	if err := rsi.computeHistoricalRSI(); err != nil {
		return nil, err
	}
	return rsi, nil
}

func (rsi *RSI[T]) computeHistoricalRSI() error {
	// 获取所有1分钟级别的K线数据
	candles := rsi.Candles()
	if len(candles) < rsi.period+1 {
		return fmt.Errorf("数据不足，计算 RSI 至少需要 %d 个价格", rsi.period+1)
	}

	// 计算每个周期内包含多少根K线，比如rsi.Bar()为15分钟，sequence.Frequency为1分钟，
	// 那么 scale = 15，即每15根K线代表一个15分钟Bar内的一个对应采样点
	scale := int(rsi.Bar() / sequence.Frequency)

	// 对于1分钟数据来说，15分钟的RSI实际可以拆分为scale个时间段，
	// 每个时间段的数据分别采样自candles中索引满足“i*scale + a”的数据，其中 a 表示偏移量(0 <= a < scale)。
	// 比如 a == 0 的时间序列对应 [0, 15, 30, 45, …] 的价格， a == 1 对应 [1, 16, 31, 46, …] 的价格，以此类推。
	// 我们分别计算这scale个时间段对应的历史RSI状态，并存入 rsi.historyAvg[a] 中。
	for a := 0; a < scale; a++ {
		// 对于每个时间段，从对应序列中提取数据来计算15分钟RSI

		// 计算初始窗口内的平均涨幅与平均跌幅
		// 例如15:00时刻的RSI，其初始计算应该使用 [0]、[15]、[30]、[45]……的价格，
		// 对应的索引即为 a, a+scale, a+2*scale, …, a + rsi.period*scale
		var gainSum, lossSum float64
		for i := 1; i <= rsi.period; i++ {
			diff := candles[i*scale+a].Value - candles[(i-1)*scale+a].Value
			if diff > 0 {
				gainSum += float64(diff)
			} else {
				lossSum += float64(-diff)
			}
		}
		avgGain := gainSum / float64(rsi.period)
		avgLoss := lossSum / float64(rsi.period)

		// 迭代更新平均涨幅和平均跌幅，采用RSI的平滑更新公式：
		// 对于每个新的Bar（15分钟），计算对应的涨跌幅，然后递推更新avgGain和avgLoss。
		// 这里循环变量 i 表示第i个15分钟Bar在该时间段内的索引位置
		for i := rsi.period + 1; i < len(candles)/scale; i++ {
			diff := candles[i*scale+a].Value - candles[(i-1)*scale+a].Value
			var currentGain, currentLoss float64
			if diff > 0 {
				currentGain = float64(diff)
				currentLoss = 0
			} else {
				currentGain = 0
				currentLoss = -float64(diff)
			}
			avgGain = (avgGain*float64(rsi.period-1) + currentGain) / float64(rsi.period)
			avgLoss = (avgLoss*float64(rsi.period-1) + currentLoss) / float64(rsi.period)
		}

		// 将当前时间段a对应的最终RSI计算状态保存到历史状态中
		rsi.historyAvg[a] = gainLoss{avgGain, avgLoss}
	}
	return nil
}

// Update Hook
// Update Hook
func (rsi *RSI[T]) Update() (*sequence.Candle[T], error) {
	// 调用底层 Sequence 的 Update() 得到最新的 1 分钟 Candle

	// 备份上一次的 RSI 值
	rsi.lastRSIQueue = append(rsi.lastRSIQueue, rsi.CurrentRSI())
	if len(rsi.lastRSIQueue) > 5 {
		rsi.lastRSIQueue = rsi.lastRSIQueue[1:]
	}

	_, err := rsi.Sequence.Update()
	if err != nil {
		return nil, err
	}

	// index 表示当前更新对应的时间序列 offset，
	// 例如如果 rsi.Bar() 为 15 分钟且 sequence.Frequency 为 1 分钟，则 scale = 15，
	// 那么 index 可能为 0、1、…、14，分别对应 15:00、15:01、…、15:14 时刻的 15 分钟 RSI 采样
	index := rsi.Sequence.LastBarIndex()

	// 获取所有的1分钟数据
	candles := rsi.Candles()
	// scale：1个Bar中包含的1分钟数据个数
	scale := int(rsi.Bar() / sequence.Frequency)

	// 当前更新的 candle 在 candles 中的下标
	lastIdx := len(candles) - 1
	// 对于同一时间序列，前一个采样点应当是 lastIdx - scale
	prevIdx := lastIdx - scale
	if prevIdx < 0 {
		// 如果数据不足一个Bar，则无法计算当前 RSI 更新，直接返回
		return nil, nil
	}

	// 计算本Bar在该时间序列内的价格差值
	diff := float64(candles[lastIdx].Value - candles[prevIdx].Value)
	var currentGain, currentLoss float64
	if diff > 0 {
		currentGain = diff
		currentLoss = 0
	} else {
		currentGain = 0
		currentLoss = -diff
	}

	// 从历史状态中取出当前时间序列的平均涨跌幅（初始值在 computeHistoricalRSI 中已计算好）
	prevAvg := rsi.historyAvg[index]

	// 使用 RSI 标准的平滑公式更新当前时间序列的平均涨幅和平均跌幅：
	// newAvg = (prevAvg*(period-1) + currentValue) / period
	newAvgGain := (prevAvg.avgGain*float64(rsi.period-1) + currentGain) / float64(rsi.period)
	newAvgLoss := (prevAvg.avgLoss*float64(rsi.period-1) + currentLoss) / float64(rsi.period)

	// 更新该时间序列对应的历史状态
	rsi.historyAvg[index] = gainLoss{avgGain: newAvgGain, avgLoss: newAvgLoss}

	return &candles[lastIdx], nil
}

func (rsi *RSI[T]) CurrentRSI() float64 {
	index := rsi.Sequence.LastBarIndex()
	avg := rsi.historyAvg[index]
	if avg.avgLoss == 0 {
		return 100
	}
	rs := avg.avgGain / avg.avgLoss
	return 100 - 100/(1+rs)
}

func (rsi *RSI[T]) LastRSIs() []float64 {
	return rsi.lastRSIQueue
}
