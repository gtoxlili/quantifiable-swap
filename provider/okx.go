package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/client"
	"github.com/gtoxlili/quantifiable-swap/common"
	"github.com/gtoxlili/quantifiable-swap/common/limiter"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"io"
	"net/http"
	"strings"
	"time"
)

type OkxProvider struct {
	latestPriceURL string
	historyURL     string
	apiKey         string
	secrectKey     string
	passphrase     string

	limiter limiter.RateLimiter

	// 下单方法
	orderFunc func(base, quote, side, size string) (string, error)
}

func NewOkx() Provider {
	return &OkxProvider{
		latestPriceURL: "https://www.okx.com/api/v5/market/index-tickers?instId=%s",
		historyURL:     "https://www.okx.com/api/v5/market/history-index-candles?instId=%s&bar=1m&limit=%d",
		apiKey:         constants.OkxAPIKey,
		secrectKey:     constants.OkxSecretKey,
		passphrase:     constants.OkxPassphrase,
		// 限速：20次/2s
		limiter: limiter.NewTokenRateLimiter(4),
	}
}

// InjectOrderFunc 注入下单方法
func (o *OkxProvider) InjectOrderFunc(orderFunc func(base, quote, side, size string) (string, error)) Provider {
	o.orderFunc = orderFunc
	return o
}

func (o *OkxProvider) MaxHistoryLimit() int {
	return 100
}

func (o *OkxProvider) Name() string {
	return "OKX"
}

func (o *OkxProvider) GetHistoryTpRes(base, quote string, afterTime string, limit int) ([]*TpRes, error) {
	if err := o.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", o.Name(), err)
	}

	url := fmt.Sprintf(o.historyURL, o.encodeInstId(base, quote), limit)
	if afterTime != "" {
		url = url + "&after=" + afterTime
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
	if response.Code != "0" {
		return nil, fmt.Errorf("接口返回错误：%s", response.Msg)
	}
	candles := response.Data
	var res []*TpRes
	for i := 0; i < len(candles); i++ {
		candle := candles[i]
		res = append(res, &TpRes{
			candle[0], // 时间
			candle[4], // 价格
		})
	}
	return res, nil
}

func (o *OkxProvider) GetLatestTpRes(base, quote string) (*TpRes, error) {
	if err := o.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", o.Name(), err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(o.latestPriceURL, o.encodeInstId(base, quote)), nil)
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
			IdxPx string `json:"idxPx"`
			Ts    string `json:"ts"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if response.Code != "0" {
		return nil, fmt.Errorf("接口返回异常：%s", response.Msg)
	}
	return &TpRes{
		response.Data[0].Ts,
		response.Data[0].IdxPx,
	}, nil
}

type OkxOrderRequest struct {
	InstId  string `json:"instId"`           // 产品ID，如 "BTC-USDT"
	TdMode  string `json:"tdMode"`           // 交易模式，现货下单固定为 "cash"
	Side    string `json:"side"`             // 订单方向：buy 或 sell
	OrdType string `json:"ordType"`          // 订单类型，这里固定为 "market"
	Sz      string `json:"sz"`               // 委托数量，市价买单时指计价币数量（默认单位为 quote_ccy）
	TgtCcy  string `json:"tgtCcy,omitempty"` // 目标币种，市价买单时指基础币种（默认单位为 base_ccy）
}

func (o *OkxProvider) MarketOrder(base, quote, side, sz string) (string, error) {

	if o.orderFunc != nil {
		return o.orderFunc(base, quote, side, sz)
	}

	// 构造请求参数，市价单时 ordType 固定为 "market"
	reqPayload := OkxOrderRequest{
		InstId:  o.encodeInstId(base, quote),
		TdMode:  "cash",
		Side:    strings.ToLower(side),
		OrdType: "market",
		Sz:      sz,
		TgtCcy:  "quote_ccy",
	}

	// 序列化请求参数为 JSON
	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}

	// 发送请求
	respReader, err := o.fetchOkxAuthRequest("POST", "/api/v5/trade/order", bodyBytes)
	if err != nil {
		return "", fmt.Errorf("request error: %v", err)
	}
	defer respReader.Close()
	var respData struct {
		Code string `json:"code"` // 结果代码，"0" 表示成功
		Msg  string `json:"msg"`  // 错误信息，成功时为空
		Data []struct {
			OrdId string `json:"ordId"`
			SCode string `json:"sCode"`
			SMsg  string `json:"sMsg"`
		} `json:"data"` // 订单数据数组
	}
	if err := json.NewDecoder(respReader).Decode(&respData); err != nil {
		return "", fmt.Errorf("json unmarshal error: %v", err)
	}
	if respData.Code != "0" {
		return "", fmt.Errorf("order failed: %s", respData.Msg)
	}
	if len(respData.Data) == 0 {
		return "", fmt.Errorf("order failed: empty data")
	}
	if respData.Data[0].SCode != "0" {
		return "", fmt.Errorf("order failed: %s", respData.Data[0].SMsg)
	}

	return respData.Data[0].OrdId, nil
}

func (o *OkxProvider) fetchOkxAuthRequest(method, requestPath string, body []byte) (io.ReadCloser, error) {
	if err := o.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("%s rate limit wait error: %v", o.Name(), err)
	}
	// 获取当前时间
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	// 生成 OK-ACCESS-SIGN
	sign, err := common.HmacSha256Sign(timestamp+method+requestPath+string(body), o.secrectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %v", err)
	}

	// 构造请求
	req, err := http.NewRequest(method, "https://www.okx.com"+requestPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OK-ACCESS-KEY", o.apiKey)
	req.Header.Set("OK-ACCESS-SIGN", base64.StdEncoding.EncodeToString(sign))
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", o.passphrase)

	resp, err := client.C.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %v", err)
	}
	return resp.Body, nil
}

func (o *OkxProvider) encodeInstId(base, quote string) string {
	return strings.ToUpper(base) + "-" + strings.ToUpper(quote)
}
