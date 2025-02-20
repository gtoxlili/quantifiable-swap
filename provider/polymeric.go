package provider

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/lo"
	"strconv"
)

type PolymericProvider struct {
	members      []Provider
	strategyFunc func([]*PriceTick) *PriceTick

	// 下单方法
	orderFunc func(base, quote, side string, size float64) (string, error)
}

func NewPolymericProvider(members ...Provider) Provider {
	return &PolymericProvider{
		members:      members,
		strategyFunc: defaultStrategy,
	}
}

func NewPolymericProviderWithStrategy(strategy func([]*PriceTick) *PriceTick, members ...Provider) Provider {
	return &PolymericProvider{
		members:      members,
		strategyFunc: strategy,
	}
}

func (p *PolymericProvider) WithOrderInjection(orderFunc func(base, quote, side string, size float64) (string, error)) Provider {
	newProvider := *p
	newProvider.orderFunc = orderFunc
	return &newProvider
}

func (p *PolymericProvider) GetHistoricalData(base, quote string, afterTime string, limit int) ([]*PriceTick, error) {
	type concurrencyResult struct {
		res []*PriceTick
		err error
	}

	results := lo.MapConcurrent(p.members, func(member Provider, _ int) concurrencyResult {
		tpRes, e := member.GetHistoricalData(base, quote, afterTime, limit)
		if e != nil {
			return concurrencyResult{
				err: fmt.Errorf("member %s: %v", member.Name(), e),
			}
		}
		return concurrencyResult{res: tpRes}
	})

	var allRes [][]*PriceTick
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		allRes = append(allRes, r.res)
	}

	var result []*PriceTick
	minLen := lo.MinLen(allRes)
	for i := 0; i < minLen; i++ {
		var tpResGroup []*PriceTick
		for j := 0; j < len(allRes); j++ {
			tpResGroup = append(tpResGroup, allRes[j][i])
		}
		result = append(result, p.strategyFunc(tpResGroup))
	}

	return result, nil
}

func (p *PolymericProvider) GetLatestData(base, quote string) (*PriceTick, error) {
	type result struct {
		res *PriceTick
		err error
	}

	results := lo.MapConcurrent(p.members, func(member Provider, _ int) result {
		tpRes, e := member.GetLatestData(base, quote)
		if e != nil {
			return result{
				err: fmt.Errorf("member %s: %v", member.Name(), e),
			}
		}
		return result{res: tpRes}
	})

	var finalRes []*PriceTick
	for _, res := range results {
		if res.err != nil {
			return nil, res.err
		}
		finalRes = append(finalRes, res.res)
	}
	return p.strategyFunc(finalRes), nil
}

func (p *PolymericProvider) ExecuteMarketOrder(base, quote string, side string, size float64) (string, error) {
	if p.orderFunc == nil {
		panic("implement me")
	}
	return p.orderFunc(base, quote, side, size)
}

func (p *PolymericProvider) GetMaxHistoryLimit() int {
	var limArr []int
	for _, member := range p.members {
		limArr = append(limArr, member.GetMaxHistoryLimit())
	}
	return lo.Min(limArr...)
}

func (p *PolymericProvider) Name() string {
	name := ""
	for _, member := range p.members {
		name += member.Name() + "|"
	}
	return name[:len(name)-1]
}

func (p *PolymericProvider) encodeInstrumentID(_, _ string) string {
	panic("undefined behavior")
}

// 默认的策略函数 （AVG）
func defaultStrategy(tpResGroup []*PriceTick) *PriceTick {
	var sum float64
	for _, tpRes := range tpResGroup {
		price, e := strconv.ParseFloat(tpRes.Price, 64)
		if e != nil {
			return nil
		}
		sum += price
	}
	avg := sum / float64(len(tpResGroup))
	return &PriceTick{
		Timestamp: tpResGroup[0].Timestamp,
		Price:     fmt.Sprintf("%f", avg),
	}
}
