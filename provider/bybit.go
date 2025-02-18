package provider

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/client"
	"github.com/gtoxlili/quantifiable-swap/common"
	"github.com/gtoxlili/quantifiable-swap/common/limiter"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ByBitProvider struct {
	latestPriceURL string
	historyURL     string
	apiKey         string
	apiSecret      string

	limiter limiter.RateLimiter

	// 下单方法
	orderFunc func(base, quote, side string, size float64) (string, error)
}

func NewByBit() Provider {
	return &ByBitProvider{
		latestPriceURL: "https://api.bybit.com/v5/market/tickers?category=spot&symbol=%s",
		historyURL:     "https://api.bybit.com/v5/market/kline?category=spot&interval=1&symbol=%s&limit=%d",
		apiKey:         constants.ByBitAPIKey,
		apiSecret:      constants.ByBitAPISecret,
		// Bybit:
		// - Limit: 600 requests in a rolling 5-second window => 120 requests per second.
		// - Example token limiter: rps = 120, burst = 120.
		limiter: limiter.NewTokenRateLimiterWithBurst(120, 120),
	}
}

// InjectOrderFunc 注入下单方法
func (b ByBitProvider) InjectOrderFunc(orderFunc func(base, quote, side string, size float64) (string, error)) Provider {
	b.orderFunc = orderFunc
	return &b
}

func (b *ByBitProvider) GetHistoryTpRes(base, quote string, afterTime string, limit int) ([]*TpRes, error) {
	if err := b.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", b.Name(), err)
	}

	url := fmt.Sprintf(b.historyURL, b.encodeInstId(base, quote), limit)
	if afterTime != "" {
		// 减一分钟
		tmpAfterTime, _ := strconv.ParseInt(afterTime, 10, 64)
		url = url + "&end=" + fmt.Sprintf("%d", tmpAfterTime-60*1000)
	} else {
		// 只拿一分钟前的数据
		url = url + "&end=" + fmt.Sprintf("%d", time.Now().Unix()*1000-60*1000)
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
	var response struct {
		RetCode int    `json:"retCode"`
		RetMsg  string `json:"retMsg"`
		Result  struct {
			List [][]string `json:"list"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if response.RetCode != 0 {
		return nil, fmt.Errorf("retCode: %d, retMsg: %s", response.RetCode, response.RetMsg)
	}
	var res []*TpRes
	for i := 0; i < len(response.Result.List); i++ {
		candle := response.Result.List[i]
		res = append(res, &TpRes{
			Timestamp: candle[0], // 时间
			Price:     candle[4], // 价格
		})
	}
	return res, nil
}

func (b *ByBitProvider) GetLatestTpRes(base, quote string) (*TpRes, error) {
	if err := b.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", b.Name(), err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(b.latestPriceURL, b.encodeInstId(base, quote)), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.C.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response struct {
		RetCode int    `json:"retCode"`
		RetMsg  string `json:"retMsg"`
		Result  struct {
			List []struct {
				LastPrice string `json:"lastPrice"`
			} `json:"list"`
		} `json:"result"`
		Time int64 `json:"time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if response.RetCode != 0 {
		return nil, fmt.Errorf("retCode: %d, retMsg: %s", response.RetCode, response.RetMsg)
	}
	if len(response.Result.List) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return &TpRes{
		Timestamp: strconv.FormatInt(response.Time, 10),
		Price:     response.Result.List[0].LastPrice,
	}, nil
}

type ByBitOrderRequest struct {
	Category    string `json:"category"`    // 例如："spot"、"linear"、"inverse"
	Symbol      string `json:"symbol"`      // 例如："BTCUSDT"
	Side        string `json:"side"`        // "Buy" 或 "Sell"
	OrderType   string `json:"orderType"`   // "Market"
	Qty         string `json:"qty"`         // 例如："0.001" (BTC) 或 "100" (USDT)
	TimeInForce string `json:"timeInForce"` // 对于市价单，可以设置为 "IOC"
	MarketUnit  string `json:"marketUnit"`
}

func (b *ByBitProvider) MarketOrder(base, quote string, side string, size float64) (string, error) {
	if b.orderFunc != nil {
		return b.orderFunc(base, quote, side, size)
	}

	reqPayload := ByBitOrderRequest{
		Category:    "spot",
		Symbol:      b.encodeInstId(base, quote),
		Side:        strings.Title(side),
		OrderType:   "Market",
		Qty:         fmt.Sprintf("%f", size),
		TimeInForce: "IOC",
		MarketUnit:  "quoteCoin",
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}

	respBody, err := b.fetchByBitAuthRequest("POST", "/v5/order/create", bodyBytes)
	if err != nil {
		return "", fmt.Errorf("request error: %v", err)
	}
	defer respBody.Close()
	var respData struct {
		RetCode int    `json:"retCode"`
		RetMsg  string `json:"retMsg"`
		Result  struct {
			OrderId string `json:"orderId"`
		} `json:"result"`
	}
	if err := json.NewDecoder(respBody).Decode(&respData); err != nil {
		return "", fmt.Errorf("json unmarshal error: %v", err)
	}
	if respData.RetCode != 0 {
		return "", fmt.Errorf("order failed: %s", respData.RetMsg)
	}
	return respData.Result.OrderId, nil
}

func (b *ByBitProvider) MaxHistoryLimit() int {
	return 1000
}

func (b *ByBitProvider) Name() string {
	return "ByBit"
}

func (b *ByBitProvider) encodeInstId(base, quote string) string {
	return strings.ToUpper(base) + strings.ToUpper(quote)
}

func (b *ByBitProvider) fetchByBitAuthRequest(method, requestPath string, body []byte) (io.ReadCloser, error) {
	if err := b.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", b.Name(), err)
	}
	// 获取当前时间戳
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	signString := timestamp + b.apiKey + string(body)
	signature, err := common.HmacSha256Sign(signString, b.apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %v", err)
	}

	// 构建请求
	url := "https://api.bybit.com" + requestPath
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-API-KEY", b.apiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-SIGN", hex.EncodeToString(signature))

	resp, err := client.C.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %v", err)
	}

	// 返回响应体
	return resp.Body, nil
}
