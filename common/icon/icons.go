package icon

import (
	"errors"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/client"
	"net/http"
	"strings"
)

// 用于存储 coin 链接的 map
var coinMap = map[string]string{
	// 特殊情况
	"AERO": "https://s1.bycsi.com/app/assets/token/459ae940b893cb3e778805ab3effcff0.png",
}

// 解析 Symbol
func parseSymbol(symbol string) string {
	s := strings.Split(symbol, "-")
	return strings.ToUpper(s[0])
}

var ErrCoinNotFound = errors.New("coin not found")

// GetCoinIcon 获取 coin 的图标链接
func GetCoinIcon(symbol string) (string, error) {
	coin := parseSymbol(symbol)
	if url, ok := coinMap[coin]; ok {
		return url, nil
	}
	// 判断 "https://bin.bnbstatic.com/static/assets/logos/{coin}.png" 是否存在
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://bin.bnbstatic.com/static/assets/logos/%s.png", coin), nil)
	resp, err := client.C.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ErrCoinNotFound
	}
	coinMap[coin] = fmt.Sprintf("https://bin.bnbstatic.com/static/assets/logos/%s.png", coin)
	return coinMap[coin], nil
}
