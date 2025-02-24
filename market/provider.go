package market

import (
	"strings"
)

type PriceTick struct {
	Timestamp string
	Price     string
}

// Provider 定义了市场提供者的基本接口
type Provider interface {
	// DataProvider 市场数据能力
	DataProvider
	// TradingProvider 交易能力
	TradingProvider
}

type ProviderInfo interface {
	Name() string
	UrlScheme(base, quote string) string
}

// DataProvider 定义了市场数据提供能力
type DataProvider interface {
	ProviderInfo
	GetHistoricalData(base, quote string, afterTime string, limit int) ([]*PriceTick, error)
	GetLatestData(base, quote string) (*PriceTick, error)
	GetMaxHistoryLimit() int
}

// TradingProvider 定义了交易能力
type TradingProvider interface {
	ProviderInfo
	ExecuteMarketOrder(base, quote string, side string, size float64) (string, error)
}

// providers 存储所有注册的 Provider 实例
var (
	dataProviders    = make(map[string]DataProvider)
	tradingProviders = make(map[string]TradingProvider)
)

func registerProvider(p any) {
	if dp, ok := p.(DataProvider); ok {
		dataProviders[strings.ToLower(dp.Name())] = dp
	}
	if tp, ok := p.(TradingProvider); ok {
		tradingProviders[strings.ToLower(tp.Name())] = tp
	}
}

// init 在包初始化时注册所有内置的 Provider
func init() {
	registerProvider(NewOkx())
	registerProvider(NewByBit())
	registerProvider(NewBinance())
	registerProvider(NewBitGet())
}

func NewDataProvider(name string) DataProvider {
	return dataProviders[strings.ToLower(name)]
}

func NewTradingProvider(name string) TradingProvider {
	return tradingProviders[strings.ToLower(name)]
}

// ListAvailableProviders returns a list of all registered provider names
func ListAvailableProviders(typ string) []string {
	var names []string
	switch typ {
	case "数据":
		for name := range dataProviders {
			names = append(names, name)
		}
	case "交易":
		for name := range tradingProviders {
			names = append(names, name)
		}
	}
	return names
}
