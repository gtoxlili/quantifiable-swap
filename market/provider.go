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
	// Name 提供者标识
	Name() string
}

// DataProvider 定义了市场数据提供能力
type DataProvider interface {
	GetHistoricalData(base, quote string, afterTime string, limit int) ([]*PriceTick, error)
	GetLatestData(base, quote string) (*PriceTick, error)
	GetMaxHistoryLimit() int
	WithOrderInjection(orderFunc func(base, quote, side string, size float64) (string, error)) Provider
}

// TradingProvider 定义了交易能力
type TradingProvider interface {
	ExecuteMarketOrder(base, quote string, side string, size float64) (string, error)
}

// providers 存储所有注册的 Provider 实例
var providers = make(map[string]Provider)

// RegisterProvider 注册一个 Provider 实例
func registerProvider(p Provider) {
	providers[strings.ToLower(p.Name())] = p
}

// init 在包初始化时注册所有内置的 Provider
func init() {
	registerProvider(NewOkx())
	registerProvider(NewByBit())
	registerProvider(NewBinance())
	registerProvider(NewBitGet())
}

// NewProvider 根据名称返回对应的 Provider
func NewProvider(name string) Provider {
	return providers[strings.ToLower(name)]
}

// ListAvailableProviders returns a list of all registered provider names
func ListAvailableProviders() []string {
	var names []string
	for name := range providers {
		names = append(names, name)
	}
	return names
}
