package provider

type TpRes struct {
	Timestamp string
	Price     string
}

// Provider 数据提供者
type Provider interface {
	GetHistoryTpRes(base, quote string, afterTime string, limit int) ([]*TpRes, error)
	GetLatestTpRes(base, quote string) (*TpRes, error)
	MarketOrder(base, quote string, side string, size string) (string, error)
	// MaxHistoryLimit 可容忍的历史数据最大获取量
	MaxHistoryLimit() int
	// Name 供应商名称
	Name() string
	// EncodeInstId 通用性设计：编码不同格式的 instId
	encodeInstId(base, quote string) string
	InjectOrderFunc(orderFunc func(base, quote, side, size string) (string, error)) Provider
}

// 一个用于存储所有实现了 Provider 接口的实例的 map
var providers = loadProviderImplementations()

func loadProviderImplementations() map[string]Provider {
	discovered := make(map[string]Provider)

	okx := NewOkx()
	bybit := NewByBit()
	binance := NewBinance()
	bitget := NewBitGet()

	discovered[okx.Name()] = okx
	discovered[bybit.Name()] = bybit
	discovered[binance.Name()] = binance
	discovered[bitget.Name()] = bitget

	return discovered
}

// 根据名称返回对应的 Provider
func NewProvider(name string) Provider {
	return providers[name]
}
