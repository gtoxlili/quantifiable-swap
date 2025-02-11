package provider

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/lo"
	"strconv"
)

type PolymericProvider struct {
	members      []Provider
	strategyFunc func([]*TpRes) *TpRes
}

func NewPolymericProvider(members ...Provider) Provider {
	return &PolymericProvider{
		members:      members,
		strategyFunc: defaultStrategy,
	}
}

func NewPolymericProviderWithStrategy(strategy func([]*TpRes) *TpRes, members ...Provider) Provider {
	return &PolymericProvider{
		members:      members,
		strategyFunc: strategy,
	}
}

func (p *PolymericProvider) GetHistoryTpRes(base, quote string, afterTime string, limit int) ([]*TpRes, error) {
	type concurrencyResult struct {
		res []*TpRes
		err error
	}

	results := lo.MapConcurrent(p.members, func(member Provider, _ int) concurrencyResult {
		tpRes, e := member.GetHistoryTpRes(base, quote, afterTime, limit)
		if e != nil {
			return concurrencyResult{
				err: fmt.Errorf("member %s: %v", member.Name(), e),
			}
		}
		return concurrencyResult{res: tpRes}
	})

	var allRes [][]*TpRes
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		allRes = append(allRes, r.res)
	}

	var result []*TpRes
	minLen := lo.MinLen(allRes)
	for i := 0; i < minLen; i++ {
		var tpResGroup []*TpRes
		for j := 0; j < len(allRes); j++ {
			tpResGroup = append(tpResGroup, allRes[j][i])
		}
		result = append(result, p.strategyFunc(tpResGroup))
	}

	return result, nil
}

func (p *PolymericProvider) GetLatestTpRes(base, quote string) (*TpRes, error) {
	type result struct {
		res *TpRes
		err error
	}

	results := lo.MapConcurrent(p.members, func(member Provider, _ int) result {
		tpRes, e := member.GetLatestTpRes(base, quote)
		if e != nil {
			return result{
				err: fmt.Errorf("member %s: %v", member.Name(), e),
			}
		}
		return result{res: tpRes}
	})

	var finalRes []*TpRes
	for _, res := range results {
		if res.err != nil {
			return nil, res.err
		}
		finalRes = append(finalRes, res.res)
	}
	return p.strategyFunc(finalRes), nil
}

func (p *PolymericProvider) MarketOrder(base, quote string, side string, size string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PolymericProvider) MaxHistoryLimit() int {
	var limArr []int
	for _, member := range p.members {
		limArr = append(limArr, member.MaxHistoryLimit())
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

func (p *PolymericProvider) encodeInstId(_, _ string) string {
	panic("undefined behavior")
}

// 默认的策略函数 （AVG）
func defaultStrategy(tpResGroup []*TpRes) *TpRes {
	var sum float64
	for _, tpRes := range tpResGroup {
		price, e := strconv.ParseFloat(tpRes.Price, 64)
		if e != nil {
			return nil
		}
		sum += price
	}
	avg := sum / float64(len(tpResGroup))
	return &TpRes{
		Timestamp: tpResGroup[0].Timestamp,
		Price:     fmt.Sprintf("%f", avg),
	}
}
