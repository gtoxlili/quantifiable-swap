package trading

import (
	"golang.org/x/exp/slices"
	"time"
)

type TradeContext struct {
	rsiQueue []float64      // 过去 5 个 RSI
	curRSI   float64        // 当前 RSI
	snapshot *TradeSnapshot // 上次交易快照（RSI, Price, OrderTime）
	bar      time.Duration  // K 线周期
	price    float64        // 当前价格
	ma5      float64        // 5 周期均线
	ma20     float64        // 20 周期均线
}

func (r TradeContext) Snapshot(snapshot *TradeSnapshot) TradeContext {
	r.snapshot = snapshot
	return r
}

// sellStrategy determines if the RSI suggests a sell signal
func defaultCanSell(tc TradeContext) error {
	rsiQueue := tc.rsiQueue
	curRSI := tc.curRSI
	snapshot := tc.snapshot
	bar := tc.bar

	// 如果包含 -1，说明数据不足，无法判断
	if slices.Contains(rsiQueue, -1) {
		return ErrInsufficientData
	}
	// 下标顺序: [0,1,2,3,4] -> older -> newer
	lastRSI := rsiQueue[4]
	lastLastRSI := rsiQueue[3]
	lastLastLastRSI := rsiQueue[2]

	// 卖出判断逻辑: up -> up -> * -> down && RSI > 70
	if lastRSI > lastLastRSI && lastLastRSI > lastLastLastRSI &&
		curRSI < lastRSI && lastRSI > 70 {
		if time.Since(snapshot.OrderTime) <= 4*bar && curRSI < 1.2*snapshot.RSI {
			// 避免短期一直在 70 附近震荡的情况
			snapshot.OrderTime = time.Now()
			return ErrFrequentSell
		}
		return nil
	}
	return ErrInvalidSellCondition
}

func defaultCanBuy(tc TradeContext) error {
	rsiQueue := tc.rsiQueue
	curRSI := tc.curRSI
	snapshot := tc.snapshot
	bar := tc.bar

	// 如果包含 -1，说明数据不足，无法判断
	if slices.Contains(rsiQueue, -1) {
		return ErrInsufficientData
	}
	// 下标顺序: [0,1,2,3,4] -> older -> newer
	lastRSI := rsiQueue[4]
	lastLastRSI := rsiQueue[3]
	lastLastLastRSI := rsiQueue[2]

	// 买入判断逻辑: down -> down -> * -> up && RSI < 30
	if lastRSI < lastLastRSI && lastLastRSI < lastLastLastRSI &&
		curRSI > lastRSI && lastRSI < 30 {
		if time.Since(snapshot.OrderTime) <= 4*bar && 1.2*curRSI > snapshot.RSI {
			// 避免短期一直在 30 附近震荡的情况
			snapshot.OrderTime = time.Now()
			return ErrFrequentBuy
		}
		return nil
	}
	return ErrInvalidBuyCondition
}

// 判断是否形成金叉或者死叉
func crossOver(ma5, ma20 float64, lastMA5, lastMA20 float64) string {
	if ma5 == -1 || ma20 == -1 {
		return ""
	}
	if lastMA5 == -1 || lastMA20 == -1 {
		return ""
	}
	if lastMA5 < lastMA20 && ma5 > ma20 {
		return "金叉"
	}
	if lastMA5 > lastMA20 && ma5 < ma20 {
		return "死叉"
	}
	return ""
}
