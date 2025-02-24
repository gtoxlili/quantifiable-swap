package market

import (
	"encoding/json"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/client"
	"github.com/gtoxlili/quantifiable-swap/common/limiter"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BinanceProvider struct {
	latestPriceURL string
	historyURL     string

	limiter limiter.RateLimiter
}

func NewBinance() *BinanceProvider {
	return &BinanceProvider{
		latestPriceURL: "https://api.binance.com/api/v3/ticker/price?symbol=%s",
		historyURL:     "https://api.binance.com/api/v3/klines?symbol=%s&interval=1m&limit=%d",
		// Binance :
		// - Kline data endpoint /api/v3/klines has weight = 2.
		// - Request weight limit: 1200 per minute => ~20 weight/sec.
		// With weight=2 for klines, it's ~10 klines requests/sec.
		// - IP connection limit: 300 per 5 min => 60 per min => 1 connection/sec (approx).
		// - Example token limiter: rps = 10, burst = 20 (conservative for klines requests).
		limiter: limiter.NewTokenRateLimiterWithBurst(10, 20),
	}
}

func (b *BinanceProvider) Name() string {
	return "Binance"
}

func (b *BinanceProvider) UrlScheme(_, _ string) string {
	return "bnc://app.binance.com/markets/markets"
}

func (b *BinanceProvider) GetMaxHistoryLimit() int {
	return 1000
}

func (b *BinanceProvider) GetHistoricalData(base, quote string, afterTime string, limit int) ([]*PriceTick, error) {
	if err := b.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", b.Name(), err)
	}

	url := fmt.Sprintf(b.historyURL, b.encodeInstrumentID(base, quote), limit)

	if afterTime != "" {
		// 减一分钟
		tmpAfterTime, _ := strconv.ParseInt(afterTime, 10, 64)
		url = url + "&endTime=" + fmt.Sprintf("%d", tmpAfterTime-60*1000)
	} else {
		// 只拿一分钟前的数据
		url = url + "&endTime=" + fmt.Sprintf("%d", time.Now().Unix()*1000-60*1000)
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
	var response [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	var res []*PriceTick
	// 币安是按照时间升序排列的（最早的在前面），所以要倒序遍历
	for i := len(response) - 1; i >= 0; i-- {
		candle := response[i]
		// float64
		timestamp := strconv.FormatFloat(candle[0].(float64), 'f', -1, 64)
		res = append(res, &PriceTick{
			Timestamp: timestamp,          // 时间
			Price:     candle[4].(string), // 价格
		})
	}
	//
	return res, nil
}

func (b *BinanceProvider) GetLatestData(base, quote string) (*PriceTick, error) {
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
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &PriceTick{
		Timestamp: fmt.Sprintf("%d", time.Now().UnixNano()/1e6),
		Price:     response.Price,
	}, nil
}

func (b *BinanceProvider) encodeInstrumentID(base, quote string) string {
	return strings.ToUpper(base) + strings.ToUpper(quote)
}
