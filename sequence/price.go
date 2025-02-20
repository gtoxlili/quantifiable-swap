package sequence

import (
	"context"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/market"
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
	dataProvider market.Provider
}

// NewPriceSequence 返回一个新的价格序列
func NewPriceSequence(ctx context.Context, base, quote string, bar time.Duration, maxLen int, dataProvider market.Provider) (Sequence[float64], error) {
	scale := int(bar / Frequency)
	ps := &PriceSequence{
		base:         base,
		quote:        quote,
		bar:          bar,
		maxLen:       maxLen * scale,
		dataProvider: dataProvider,
	}
	// 初始化历史数据
	if err := ps.initHistory(ctx); err != nil {
		return nil, fmt.Errorf("初始化历史数据失败：%v", err)
	}
	return ps, nil
}

func (ps *PriceSequence) initHistory(ctx context.Context) error {
	// 请求历史数据 (bar * maxLen) (因为获取的时间精度是 1m 的，所以这里的 bar 如果是 15m，那么就是 15*maxLen)
	aft := ""
	var tmpTicks []*market.PriceTick

	curLim := ps.dataProvider.GetMaxHistoryLimit()
	if curLim > ps.maxLen {
		curLim = ps.maxLen
	}

outer:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			candles, err := ps.dataProvider.GetHistoricalData(ps.base, ps.quote, aft, curLim)
			if err != nil {
				return err
			}
			if len(candles) == 0 {
				break outer
			}
			tmpTicks = append(tmpTicks, candles...)
			if len(tmpTicks) >= ps.maxLen {
				// 截掉后面多余的数据
				tmpTicks = tmpTicks[:ps.maxLen]
				break outer
			}
			aft = candles[len(candles)-1].Timestamp
		}
	}

	// 逆序追加
	for i := len(tmpTicks) - 1; i >= 0; i-- {
		timestamp, _ := strconv.ParseInt(tmpTicks[i].Timestamp, 10, 64)
		price, _ := strconv.ParseFloat(tmpTicks[i].Price, 64)
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
func (ps *PriceSequence) Update(ctx context.Context) (*Candle[float64], error) {
	// 延迟 bar 时间
	if err := ps.delay(ctx); err != nil {
		return nil, err
	}
	priceTick, err := ps.dataProvider.GetLatestData(ps.base, ps.quote)
	if err != nil {
		return nil, err
	}
	price, _ := strconv.ParseFloat(priceTick.Price, 64)
	timestamp, _ := strconv.ParseInt(priceTick.Timestamp, 10, 64)
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

func (ps *PriceSequence) delay(ctx context.Context) error {
	next := time.Now().Truncate(Frequency).Add(Frequency)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Until(next)):
		return nil
	}
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
