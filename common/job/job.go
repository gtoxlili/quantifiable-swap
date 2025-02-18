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
func (m *Manager) AddJob(j config.Job) (id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// id 不可重复
	if _, found := m.jobs[j.Id]; found {
		return "", fmt.Errorf("job %s 已存在", j.Id)
	}

	prov := provider.NewProvider(j.Provider.Name)
	if j.Provider.InjectOrder != "" {
		injectProv := provider.NewProvider(j.Provider.InjectOrder)
		prov = prov.InjectOrderFunc(injectProv.MarketOrder)
	}

	bar, err := time.ParseDuration(j.Bar)
	if err != nil {
		return "", fmt.Errorf("非法的 K 线周期: %s", j.Bar)
	}

	switch j.Type {
	case "notify":
		m.jobs[j.Id] = swap.NewNotify(j.Symbol.Base, j.Symbol.Quote, bar, prov)
	case "swap":
		m.jobs[j.Id] = swap.NewWaper(j.Symbol.Base, j.Symbol.Quote, bar, j.Amount.Sell, j.Amount.Buy, prov)
	default:
		return "", fmt.Errorf("未知的 Job 类型: %s", j.Type)
	}

	return j.Id, nil
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
