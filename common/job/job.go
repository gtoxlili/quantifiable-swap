package job

import (
	"context"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/common/lo"
	"github.com/gtoxlili/quantifiable-swap/common/smap"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"golang.org/x/exp/slices"
	"time"

	"github.com/gtoxlili/quantifiable-swap/provider"
	"github.com/gtoxlili/quantifiable-swap/swap"
)

type IManager interface {
	AddJob(j config.Job) (string, error)
	RemoveJob(id string) error
	RemoveSubscriber(id string, chatID int64) error
	RemoveAll()
	RunJob(id string) error
	StopJob(id string) error
	JobsData(subId int64) []config.Job
	JobsAllData() []config.Job
	IsRunning(id string) bool
}

// Job Map Value 的结构
type Job struct {
	conf   *config.Job
	waper  swap.IIndicatorWaper
	cancel context.CancelFunc
}

// Manager 用于管理 Job 的添加、删除和执行
type Manager struct {
	jobs smap.SyncMap[string, *Job]
}

// NewManager 创建一个新的 Manager
func NewManager() IManager {
	return &Manager{}
}

// AddJob 添加一个新的 Job
func (m *Manager) AddJob(j config.Job) (string, error) {
	if err := j.Validate("InjectOrder", "Sell", "Buy", "Subscribers"); err != nil {
		return "", err
	}

	// id 重复时，添加订阅者
	if job, found := m.jobs.Load(j.GetId()); found {
		subscribers := job.conf.Subscribers
		if slices.Contains(subscribers, j.Subscribers[0]) {
			return "", fmt.Errorf("job %s 已存在", j.GetId())
		}
		subscribers = append(subscribers, j.Subscribers[0])
		job.conf.Subscribers = subscribers
		job.waper.WithSubscribers(subscribers)
		return j.GetId(), nil
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

	var waper swap.IIndicatorWaper
	switch j.Type {
	case "notify":
		waper = swap.NewNotify(j.Symbol.Base, j.Symbol.Quote, bar, prov)
	case "swap":
		waper = swap.NewWaper(j.Symbol.Base, j.Symbol.Quote, bar, j.Amount.Sell, j.Amount.Buy, prov)
	default:
		return "", fmt.Errorf("未知的 Job 类型: %s", j.Type)
	}

	// 鲁棒性处理
	if len(j.Subscribers) == 0 {
		if constants.TGChatID != 0 {
			j.Subscribers = []int64{constants.TGChatID}
		}
	}

	waper.WithSubscribers(j.Subscribers)
	m.jobs.Store(j.GetId(), &Job{
		conf:  &j,
		waper: waper,
	})

	return j.GetId(), nil
}

// RemoveJob 根据 id 删除 Job
func (m *Manager) RemoveJob(id string) error {
	if job, found := m.jobs.Load(id); found {
		if job.cancel != nil {
			if err := m.StopJob(id); err != nil {
				return err
			}
		}
		m.jobs.Delete(id)
		return nil
	}
	return fmt.Errorf("job %s 不存在", id)
}

// RemoveSubscriber 删除某个订阅者，如果没有订阅者则删除 Job
func (m *Manager) RemoveSubscriber(id string, chatID int64) error {
	if job, found := m.jobs.Load(id); found {
		subscribers := lo.Delete(job.conf.Subscribers, chatID)
		if len(subscribers) == 0 {
			return m.RemoveJob(id)
		}
		job.conf.Subscribers = subscribers
		job.waper.WithSubscribers(subscribers)
		return nil
	}
	return fmt.Errorf("job %s 不存在", id)
}

// RemoveAll 删除所有 Job
func (m *Manager) RemoveAll() {
	m.jobs.Range(func(key string, value *Job) bool {
		_ = m.RemoveJob(key)
		return true
	})
}

// RunJob 根据 id 运行 Job
func (m *Manager) RunJob(id string) error {
	if job, found := m.jobs.Load(id); found {
		if job.cancel != nil {
			return fmt.Errorf("job %s 已在运行", id)
		}
		ctx, cancel := context.WithCancel(context.Background())
		job.cancel = cancel
		go func() {
			job.waper.Run(ctx)
			_ = m.StopJob(id)
		}()
		return nil
	}
	return fmt.Errorf("job %s 不存在", id)
}

// StopJob 停止某个 Job
func (m *Manager) StopJob(id string) error {
	if job, found := m.jobs.Load(id); found {
		if job.cancel == nil {
			return fmt.Errorf("job %s 未在运行", id)
		}
		job.cancel()
		job.cancel = nil
		return nil
	}
	return fmt.Errorf("job %s 不存在", id)
}

// JobsAllData 获取 所有 []Job
func (m *Manager) JobsAllData() []config.Job {
	var jobs []config.Job
	m.jobs.Range(func(key string, value *Job) bool {
		jobs = append(jobs, *value.conf)
		return true
	})
	return jobs
}

// JobsData 获取某个订阅者的 []Job
func (m *Manager) JobsData(subId int64) []config.Job {
	var jobs []config.Job
	m.jobs.Range(func(key string, value *Job) bool {
		if slices.Contains(value.conf.Subscribers, subId) {
			jobs = append(jobs, *value.conf)
		}
		return true
	})
	return jobs
}

// IsRunning 判断某个 Job 是否在运行
func (m *Manager) IsRunning(id string) bool {
	if job, found := m.jobs.Load(id); found {
		return job.cancel != nil
	}
	return false
}
