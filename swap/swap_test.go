package swap

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/provider"
	"github.com/gtoxlili/quantifiable-swap/sequence"
	"golang.org/x/tools/container/intsets"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCanSell(t *testing.T) {
	rsiQueue := []float64{33, 29.5, // 降到<30，触发买入
		40, 65, 75, // 升到>70，触发卖出
	}
	curRSI := 28.0
	lastBuyTrade := &LastTrade{
		OrderTime: time.Unix(0, 0),
		RSI:       float64(intsets.MaxInt),
		Price:     float64(intsets.MaxInt),
	}

	bar := time.Minute

	err := defaultCanSell(TradeCondition{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		lst:      lastBuyTrade,
		bar:      bar,
	})
	assert.Nil(t, err)
}

func TestDefaultCanSell_NotEnoughData(t *testing.T) {
	rsiQueue := []float64{60, 65, -1, 75, 80}
	curRSI := 65.0
	lastTrade := &LastTrade{OrderTime: time.Unix(0, 0)}
	bar := time.Minute

	err := defaultCanSell(TradeCondition{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		lst:      lastTrade,
		bar:      bar,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "采样数据不足", err.Error())
}

func TestDefaultCanSell_NotMeetingSellCondition(t *testing.T) {
	rsiQueue := []float64{60, 65, 70, 75, 80}
	curRSI := 85.0
	lastTrade := &LastTrade{OrderTime: time.Unix(0, 0)}
	bar := time.Minute

	err := defaultCanSell(TradeCondition{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		lst:      lastTrade,
		bar:      bar,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "尚不满足卖出条件", err.Error())
}

func TestDefaultCanSell_SellTooFrequent(t *testing.T) {
	rsiQueue := []float64{60, 65, 70, 75, 80}
	curRSI := 65.0

	lastTrade := &LastTrade{
		OrderTime: time.Now(),
		// rsi 取一个极大值
		RSI:   float64(intsets.MaxInt),
		Price: float64(intsets.MaxInt),
	}
	bar := time.Minute

	// time.Since(lst.OrderTime) <= 4*bar && curRSI < 1.2*lst.RSI
	assert.True(t, time.Since(lastTrade.OrderTime) <= 4*bar)
	assert.True(t, curRSI < 1.2*lastTrade.RSI)

	err := defaultCanSell(TradeCondition{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		lst:      lastTrade,
		bar:      bar,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "卖出太频繁", err.Error())
}

func TestDefaultCanBuy_BuyTooFrequent(t *testing.T) {
	rsiQueue := []float64{30, 25, 20, 15, 10}
	curRSI := 20.0

	lastTrade := &LastTrade{
		OrderTime: time.Now(),
		RSI:       float64(intsets.MinInt),
		Price:     float64(intsets.MinInt),
	}
	bar := time.Minute

	// time.Since(lst.OrderTime) <= 4*bar && 1.2*curRSI > lst.RSI
	assert.True(t, time.Since(lastTrade.OrderTime) <= 4*bar)
	assert.True(t, 1.2*curRSI > lastTrade.RSI)

	err := defaultCanBuy(TradeCondition{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		lst:      lastTrade,
		bar:      bar,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "买入太频繁", err.Error())
}

// mockRSIHook 用来模拟一个能返回固定序列RSI值的“假”RSI包装器
type mockRSIHook struct {
	rsiValues    []float64
	currentIndex int
}

// Update 模拟获取最新蜡烛数据，这里仅返回一个虚拟蜡烛（在 runRSILoop 中只关心时间和价格就可）
func (m *mockRSIHook) Update() (*sequence.Candle[float64], error) {
	if m.currentIndex >= len(m.rsiValues) {
		time.Sleep(1 * time.Second)
		return nil, fmt.Errorf("数据已经用完")
	}
	candle := &sequence.Candle[float64]{
		Time:  time.Now(),
		Value: 100.0,
	}
	m.currentIndex++
	return candle, nil
}

// CurrentRSI 返回当前 RSI 值
func (m *mockRSIHook) CurrentRSI() float64 {
	if m.currentIndex == 0 {
		return 50.0 // 如果还没 Update，就给个中间值
	}
	if m.currentIndex > len(m.rsiValues) {
		return 50.0
	}
	return m.rsiValues[m.currentIndex-1]
}

// LastRSIs 返回最近若干个 RSI 值
func (m *mockRSIHook) PreviousRSIs() []float64 {
	return m.rsiValues[m.currentIndex-6 : m.currentIndex-1]
}

func TestRSIWaper_runRSILoop_BuySell(t *testing.T) {
	// 设定 RSI 序列：手动让它依次经过 <30, >70 等范围
	rsiSequence := []float64{
		50, 50, 50, 50, 50, // 5 个 -1，表示数据不足
		35, 33, 29.5, // 降到<30，触发买入
		40, 65, 75, // 升到>70，触发卖出
		28, 72, // 再次降到<30 再升到>70
	}

	// 初始化一个 mockRSIHook
	mockHook := &mockRSIHook{
		rsiValues:    rsiSequence,
		currentIndex: 5,
	}

	okxProvider := provider.NewOkx()
	r := NewRSIWaper("BTC", "USDT", time.Minute, "10", "10", okxProvider)

	go func() {
		time.Sleep(2 * time.Second)
		r.Stop()
	}()

	r.runRSILoop(mockHook)
}
