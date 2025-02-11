package sequence

import (
	"fmt"
	"quantifiable-swap/provider"
	"strconv"
	"sync"
	"time"
)

const (
	Frequency = 1 * time.Minute
)

// PriceSequence 用于维护一段价格历史，线程安全
type PriceSequence struct {
	base      string
	quote     string
	bar       time.Duration
	timePrice []Candle[float64]
	mu        sync.Mutex
	maxLen    int
	// 数据提供商
	dataProvider provider.Provider
}

// NewPriceSequence 返回一个新的价格序列
func NewPriceSequence(base, quote string, bar time.Duration, maxLen int, dataProvider provider.Provider) (Sequence[float64], error) {
	scale := int(bar / time.Minute)
	ps := &PriceSequence{
		base:         base,
		quote:        quote,
		bar:          bar,
		maxLen:       maxLen * scale,
		dataProvider: dataProvider,
	}
	// 初始化历史数据
	if err := ps.initHistory(); err != nil {
		return nil, fmt.Errorf("初始化历史数据失败：%v", err)
	}
	return ps, nil
}

func (ps *PriceSequence) initHistory() error {
	// 请求历史数据 (bar * maxLen) (因为获取的时间精度是 1m 的，所以这里的 bar 如果是 15m，那么就是 15*maxLen)
	aft := ""
	var tmpCandles []*provider.TpRes

	curLim := ps.dataProvider.MaxHistoryLimit()
	if curLim > ps.maxLen {
		curLim = ps.maxLen
	}

	for {
		//candles, err := ps.appendTimeRange(aft, 100)
		candles, err := ps.dataProvider.GetHistoryTpRes(ps.base, ps.quote, aft, curLim)
		if err != nil {
			return err
		}
		if len(candles) == 0 {
			break
		}
		tmpCandles = append(tmpCandles, candles...)
		if len(tmpCandles) >= ps.maxLen {
			// 截掉后面多余的数据
			tmpCandles = tmpCandles[:ps.maxLen]
			break
		}
		aft = candles[len(candles)-1].Timestamp
		// 限速：10次/2s
		time.Sleep(200 * time.Millisecond)
	}
	// 逆序追加
	for i := len(tmpCandles) - 1; i >= 0; i-- {
		timestamp, _ := strconv.ParseInt(tmpCandles[i].Timestamp, 10, 64)
		price, _ := strconv.ParseFloat(tmpCandles[i].Price, 64)
		// 因为后续更新逻辑中，传入的是「实时时间」，所以这里需要将时间滞后 Frequency
		// 比如 15:02:00 的价格，应该作为 15:01:00 的收盘价
		ps.append(price, time.Unix(timestamp/1000, 0).Add(Frequency))
	}
	return nil
}

func (ps *PriceSequence) Bar() time.Duration {
	return ps.bar
}

// Update 更新价格序列
func (ps *PriceSequence) Update() (*Candle[float64], error) {
	// 延迟 bar 时间
	if err := ps.delay(); err != nil {
		return nil, err
	}
	tpPair, err := ps.dataProvider.GetLatestTpRes(ps.base, ps.quote)
	if err != nil {
		return nil, err
	}
	price, _ := strconv.ParseFloat(tpPair.Price, 64)
	timestamp, _ := strconv.ParseInt(tpPair.Timestamp, 10, 64)
	timeUnix := time.Unix(timestamp/1000, 0)
	ps.append(price, timeUnix)
	return &ps.timePrice[len(ps.timePrice)-1], nil
}

func (ps *PriceSequence) Candles() []Candle[float64] {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.timePrice
}

// LastBarIndex 获取 candles 中最后一个时间属于bar 的第几个时间段
func (ps *PriceSequence) LastBarIndex() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if len(ps.timePrice) == 0 {
		return -1
	}
	lastTime := ps.timePrice[len(ps.timePrice)-1].Time
	// 用 lastTime 减去 lastTime.Truncate(ps.bar) 得到的时间差，再除以 frequency 得到的就是时间段
	return int(lastTime.Sub(lastTime.Truncate(ps.bar)) / Frequency)
}

func (ps *PriceSequence) delay() error {
	next := time.Now().Truncate(Frequency).Add(Frequency)
	time.Sleep(time.Until(next))
	return nil
}

// Append 向价格序列追加一个新的价格，若超过 maxLen 则移除最旧的价格
func (ps *PriceSequence) append(price float64, timestamp time.Time) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if len(ps.timePrice) >= ps.maxLen {
		ps.timePrice = ps.timePrice[1:]
	}
	// 由于获取的时间可能是 15:01:59 或者 15:02:01 这种，
	// 所以需要先四舍五入将其视作为 15:02:00 的时间
	// 再作为 15:01:00 的收盘价
	ps.timePrice = append(ps.timePrice, Candle[float64]{Time: timestamp.Round(Frequency).Add(-Frequency), Value: price})
}
