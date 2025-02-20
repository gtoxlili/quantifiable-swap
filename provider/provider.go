package provider

import (
	"strings"
)

type PriceTick struct {
	Timestamp string
	Price     string
}

// Provider 数据提供者
type Provider interface {
	GetHistoricalData(base, quote string, afterTime string, limit int) ([]*PriceTick, error)
	GetLatestData(base, quote string) (*PriceTick, error)
	ExecuteMarketOrder(base, quote string, side string, size float64) (string, error)
	// GetMaxHistoryLimit 可容忍的历史数据最大获取量
	GetMaxHistoryLimit() int
	// Name 供应商名称
	Name() string
	encodeInstrumentID(base, quote string) string
	WithOrderInjection(orderFunc func(base, quote, side string, size float64) (string, error)) Provider
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
