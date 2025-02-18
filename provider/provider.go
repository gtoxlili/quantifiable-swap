package provider

import "strings"

type TpRes struct {
	Timestamp string
	Price     string
}

// Provider 数据提供者
type Provider interface {
	GetHistoryTpRes(base, quote string, afterTime string, limit int) ([]*TpRes, error)
	GetLatestTpRes(base, quote string) (*TpRes, error)
	MarketOrder(base, quote string, side string, size float64) (string, error)
	// MaxHistoryLimit 可容忍的历史数据最大获取量
	MaxHistoryLimit() int
	// Name 供应商名称
	Name() string
	// EncodeInstId 通用性设计：编码不同格式的 instId
	encodeInstId(base, quote string) string
	InjectOrderFunc(orderFunc func(base, quote, side string, size float64) (string, error)) Provider
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
	return providers[name]
}
