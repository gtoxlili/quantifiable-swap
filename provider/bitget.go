package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"quantifiable-swap/client"
	"strings"
	"time"
)

type BitGetProvider struct {
	latestPriceURL string
	historyURL     string
}

func NewBitGet() Provider {
	return &BitGetProvider{
		latestPriceURL: "https://api.bitget.com/api/v2/spot/market/tickers?symbol=%s",
		historyURL:     "https://api.bitget.com/api/v2/spot/market/history-candles?symbol=%s&granularity=1min&limit=%d",
	}
}

func (b *BitGetProvider) Name() string {
	return "BitGet"
}

func (b *BitGetProvider) MaxHistoryLimit() int {
	return 200
}

func (b *BitGetProvider) GetHistoryTpRes(base, quote string, afterTime string, limit int) ([]*TpRes, error) {
	url := fmt.Sprintf(b.historyURL, b.encodeInstId(base, quote), limit)
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

func (b *BitGetProvider) GetLatestTpRes(base, quote string) (*TpRes, error) {
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
	return &TpRes{
		response.Data[0].Ts,
		response.Data[0].LastPr,
	}, nil
}

func (b *BitGetProvider) MarketOrder(base, quote string, side string, size string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BitGetProvider) encodeInstId(base, quote string) string {
	return strings.ToUpper(base) + strings.ToUpper(quote)
}
