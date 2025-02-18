package job

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"sync"
	"time"

	"github.com/gtoxlili/quantifiable-swap/provider"
	"github.com/gtoxlili/quantifiable-swap/swap"
)

// Manager 用于管理 Job 的添加、删除和执行
type Manager struct {
	mu   sync.Mutex
	jobs map[string]swap.IndicatorJob
}

// NewManager 创建一个新的 Manager
func NewManager() *Manager {
	return &Manager{
		jobs: make(map[string]swap.IndicatorJob),
	}
}

// AddJob 添加一个新的 Job
func (m *Manager) AddJob(j config.Job) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := j.Validate("Id", "InjectOrder", "Sell", "Buy"); err != nil {
		return "", err
	}

	// id 不可重复
	if _, found := m.jobs[j.GetId()]; found {
		return "", fmt.Errorf("job %s 已存在", j.GetId())
	}

	prov := provider.NewProvider(j.Provider.Name)
	if prov == nil {
		return "", fmt.Errorf("未知的 Provider: %s", j.Provider.Name)
	}
	if j.Provider.InjectOrder != "" {
		injectProv := provider.NewProvider(j.Provider.InjectOrder)
		if injectProv == nil {
			return "", fmt.Errorf("未知的 InjectOrder Provider: %s", j.Provider.InjectOrder)
		}
		prov = prov.InjectOrderFunc(injectProv.MarketOrder)
	}

	bar, err := time.ParseDuration(j.Bar)
	if err != nil {
		return "", fmt.Errorf("非法的 K 线周期: %s", j.Bar)
	}

	switch j.Type {
	case "notify":
		m.jobs[j.GetId()] = swap.NewNotify(j.Symbol.Base, j.Symbol.Quote, bar, prov)
	case "swap":
		m.jobs[j.GetId()] = swap.NewWaper(j.Symbol.Base, j.Symbol.Quote, bar, j.Amount.Sell, j.Amount.Buy, prov)
	default:
		return "", fmt.Errorf("未知的 Job 类型: %s", j.Type)
	}

	return j.GetId(), nil
}

// RemoveJob 根据 id 删除 Job
func (m *Manager) RemoveJob(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, found := m.jobs[id]; found {
		job.Stop()
		delete(m.jobs, id)
	}
	return fmt.Errorf("job %s 不存在", id)
}

// RemoveAll 终止所有 Job
func (m *Manager) RemoveAll() {
	for id, _ := range m.jobs {
		_ = m.RemoveJob(id)
	}
}

// RunJob 根据 id 运行 Job
func (m *Manager) RunJob(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, found := m.jobs[id]; found {
		go job.Run()
		return nil
	}
	return fmt.Errorf("job %s 不存在", id)
}

// JobNames 获取所有 Job Name
func (m *Manager) JobNames() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var names []string
	for id, _ := range m.jobs {
		names = append(names, id)
	}
	return names
}
