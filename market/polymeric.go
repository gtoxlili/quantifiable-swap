package market

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/lo"
	"strconv"
)

type PolymericProvider struct {
	members      []DataProvider
	strategyFunc func([]*PriceTick) *PriceTick
}

func NewPolymericProvider(members ...DataProvider) DataProvider {
	return &PolymericProvider{
		members:      members,
		strategyFunc: defaultStrategy,
	}
}

func NewPolymericProviderWithStrategy(strategy func([]*PriceTick) *PriceTick, members ...DataProvider) DataProvider {
	return &PolymericProvider{
		members:      members,
		strategyFunc: strategy,
	}
}

func (p *PolymericProvider) GetHistoricalData(base, quote string, afterTime string, limit int) ([]*PriceTick, error) {
	type concurrencyResult struct {
		res []*PriceTick
		err error
	}

	results := lo.MapConcurrent(p.members, func(member DataProvider, _ int) concurrencyResult {
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

	results := lo.MapConcurrent(p.members, func(member DataProvider, _ int) result {
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
