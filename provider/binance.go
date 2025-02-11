package provider

import (
	"encoding/json"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/client"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BinanceProvider struct {
	latestPriceURL string
	historyURL     string
}

func NewBinance() Provider {
	return &BinanceProvider{
		latestPriceURL: "https://api.binance.com/api/v3/ticker/price?symbol=%s",
		historyURL:     "https://api.binance.com/api/v3/klines?symbol=%s&interval=1m&limit=%d",
	}
}

func (b *BinanceProvider) Name() string {
	return "Binance"
}

func (b *BinanceProvider) MaxHistoryLimit() int {
	return 1000
}

func (b *BinanceProvider) GetHistoryTpRes(base, quote string, afterTime string, limit int) ([]*TpRes, error) {
	url := fmt.Sprintf(b.historyURL, b.encodeInstId(base, quote), limit)

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
	var res []*TpRes
	// 币安是按照时间升序排列的（最早的在前面），所以要倒序遍历
	for i := len(response) - 1; i >= 0; i-- {
		candle := response[i]
		// float64
		timestamp := strconv.FormatFloat(candle[0].(float64), 'f', -1, 64)
		res = append(res, &TpRes{
			Timestamp: timestamp,          // 时间
			Price:     candle[4].(string), // 价格
		})
	}
	//
	return res, nil
}

func (b *BinanceProvider) GetLatestTpRes(base, quote string) (*TpRes, error) {
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
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &TpRes{
		Timestamp: fmt.Sprintf("%d", time.Now().UnixNano()/1e6),
		Price:     response.Price,
	}, nil
}

func (b *BinanceProvider) MarketOrder(base, quote string, side string, size string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BinanceProvider) encodeInstId(base, quote string) string {
	return strings.ToUpper(base) + strings.ToUpper(quote)
}
