package trading

import (
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
	lastBuyTrade := &TradeSnapshot{
		OrderTime: time.Unix(0, 0),
		RSI:       float64(intsets.MaxInt),
		Price:     float64(intsets.MaxInt),
	}

	bar := time.Minute

	err := defaultCanSell(TradeContext{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		snapshot: lastBuyTrade,
		bar:      bar,
	})
	assert.Nil(t, err)
}

func TestDefaultCanSell_NotEnoughData(t *testing.T) {
	rsiQueue := []float64{60, 65, -1, 75, 80}
	curRSI := 65.0
	lastTrade := &TradeSnapshot{OrderTime: time.Unix(0, 0)}
	bar := time.Minute

	err := defaultCanSell(TradeContext{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		snapshot: lastTrade,
		bar:      bar,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "采样数据不足", err.Error())
}

func TestDefaultCanSell_NotMeetingSellCondition(t *testing.T) {
	rsiQueue := []float64{60, 65, 70, 75, 80}
	curRSI := 85.0
	lastTrade := &TradeSnapshot{OrderTime: time.Unix(0, 0)}
	bar := time.Minute

	err := defaultCanSell(TradeContext{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		snapshot: lastTrade,
		bar:      bar,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "尚不满足卖出条件", err.Error())
}

func TestDefaultCanSell_SellTooFrequent(t *testing.T) {
	rsiQueue := []float64{60, 65, 70, 75, 80}
	curRSI := 65.0

	lastTrade := &TradeSnapshot{
		OrderTime: time.Now(),
		// rsi 取一个极大值
		RSI:   float64(intsets.MaxInt),
		Price: float64(intsets.MaxInt),
	}
	bar := time.Minute

	// time.Since(snapshot.OrderTime) <= 4*bar && curRSI < 1.2*snapshot.RSI
	assert.True(t, time.Since(lastTrade.OrderTime) <= 4*bar)
	assert.True(t, curRSI < 1.2*lastTrade.RSI)

	err := defaultCanSell(TradeContext{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		snapshot: lastTrade,
		bar:      bar,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "卖出太频繁", err.Error())
}

func TestDefaultCanBuy_BuyTooFrequent(t *testing.T) {
	rsiQueue := []float64{30, 25, 20, 15, 10}
	curRSI := 20.0

	lastTrade := &TradeSnapshot{
		OrderTime: time.Now(),
		RSI:       float64(intsets.MinInt),
		Price:     float64(intsets.MinInt),
	}
	bar := time.Minute

	// time.Since(snapshot.OrderTime) <= 4*bar && 1.2*curRSI > snapshot.RSI
	assert.True(t, time.Since(lastTrade.OrderTime) <= 4*bar)
	assert.True(t, 1.2*curRSI > lastTrade.RSI)

	err := defaultCanBuy(TradeContext{
		rsiQueue: rsiQueue,
		curRSI:   curRSI,
		snapshot: lastTrade,
		bar:      bar,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "买入太频繁", err.Error())
}
