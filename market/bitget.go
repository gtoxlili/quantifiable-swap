package market

import (
	"encoding/json"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/client"
	"github.com/gtoxlili/quantifiable-swap/common/limiter"
	"net/http"
	"strings"
	"time"
)

type BitGetProvider struct {
	latestPriceURL string
	historyURL     string

	limiter limiter.RateLimiter

	// 下单方法
	orderFunc func(base, quote, side string, size float64) (string, error)
}

func NewBitGet() Provider {
	return &BitGetProvider{
		latestPriceURL: "https://api.bitget.com/api/v2/spot/market/tickers?symbol=%s",
		historyURL:     "https://api.bitget.com/api/v2/spot/market/history-candles?symbol=%s&granularity=1min&limit=%d",
		// Bitget:
		// - Limit: 20 requests per second (IP).
		// - Example token limiter: rps = 20, burst = 20.
		limiter: limiter.NewTokenRateLimiterWithBurst(20, 20),
	}
}

// WithOrderInjection 注入下单方法
func (b *BitGetProvider) WithOrderInjection(orderFunc func(base, quote, side string, size float64) (string, error)) Provider {
	newProvider := *b
	newProvider.orderFunc = orderFunc
	return &newProvider
}

func (b *BitGetProvider) Name() string {
	return "BitGet"
}

func (b *BitGetProvider) GetMaxHistoryLimit() int {
	return 200
}

func (b *BitGetProvider) GetHistoricalData(base, quote string, afterTime string, limit int) ([]*PriceTick, error) {
	if err := b.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", b.Name(), err)
	}

	url := fmt.Sprintf(b.historyURL, b.encodeInstrumentID(base, quote), limit)
	if afterTime != "" {
		url = url + "&endTime=" + afterTime
	} else {
		// 毫秒级
		url = url + "&endTime=" + fmt.Sprintf("%d000", time.Now().Unix())
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.C.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// 解析 JSON
	var response struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if response.Code != "00000" {
		return nil, fmt.Errorf("接口返回错误：%s", response.Msg)
	}
	candles := response.Data
	var res []*PriceTick
	for i := 0; i < len(candles); i++ {
		candle := candles[i]
		res = append(res, &PriceTick{
			candle[0], // 时间
			candle[4], // 价格
		})
	}
	return res, nil
}

func (b *BitGetProvider) GetLatestData(base, quote string) (*PriceTick, error) {
	if err := b.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", b.Name(), err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(b.latestPriceURL, b.encodeInstrumentID(base, quote)), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.C.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			LastPr string `json:"lastPr"`
			Ts     string `json:"ts"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if response.Code != "00000" {
		return nil, fmt.Errorf("接口返回异常：%s", response.Msg)
	}
	return &PriceTick{
		response.Data[0].Ts,
		response.Data[0].LastPr,
	}, nil
}

func (b *BitGetProvider) ExecuteMarketOrder(base, quote string, side string, size float64) (string, error) {
	if b.orderFunc == nil {
		panic("implement me")
	}
	return b.orderFunc(base, quote, side, size)
}

func (b *BitGetProvider) encodeInstrumentID(base, quote string) string {
	return strings.ToUpper(base) + strings.ToUpper(quote)
}
