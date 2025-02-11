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
}
