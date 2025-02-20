package job

import (
	"context"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/common/lo"
	"github.com/gtoxlili/quantifiable-swap/common/smap"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"github.com/gtoxlili/quantifiable-swap/market"
	"github.com/gtoxlili/quantifiable-swap/trading"
	"golang.org/x/exp/slices"
	"time"
)

// Job Map Value 的结构
type Job struct {
	conf     *config.Job
	executor trading.IStrategyExecutor
	cancel   context.CancelFunc
}

// Manager 用于管理 Job 的添加、删除和执行
type Manager struct {
	jobs smap.SyncMap[string, *Job]
}

// NewManager 创建一个新的 Manager
func NewManager() IManager {
	return &Manager{}
}

func (m *Manager) CreateJob(j config.Job) (string, error) {
	if err := j.Validate("InjectOrder", "Sell", "Buy", "Subscribers"); err != nil {
		return "", err
	}

	// id 重复时，添加订阅者
	if job, found := m.jobs.Load(j.String()); found {
		subscribers := job.conf.Subscribers
		if slices.Contains(subscribers, j.Subscribers[0]) {
			return "", fmt.Errorf("job %s 已存在", j.String())
		}
		subscribers = append(subscribers, j.Subscribers[0])
		job.conf.Subscribers = subscribers
		job.executor.WithSubscribers(subscribers)
		return j.String(), nil
	}

	prov := market.NewProvider(j.Provider.Name)
	if prov == nil {
		return "", fmt.Errorf("未知的 Provider: %s", j.Provider.Name)
	}
	if j.Provider.InjectOrder != "" {
		injectProv := market.NewProvider(j.Provider.InjectOrder)
		if injectProv == nil {
			return "", fmt.Errorf("未知的 InjectOrder Provider: %s", j.Provider.InjectOrder)
		}
		prov = prov.WithOrderInjection(injectProv.ExecuteMarketOrder)
	}

	bar, err := time.ParseDuration(j.Bar)
	if err != nil {
		return "", fmt.Errorf("非法的 K 线周期: %s", j.Bar)
	}

	var executor trading.IStrategyExecutor
	switch j.Type {
	case "monitor":
		executor = trading.NewMonitor(j.Symbol.Base, j.Symbol.Quote, bar, prov)
	case "trader":
		executor = trading.NewTrader(j.Symbol.Base, j.Symbol.Quote, bar, j.Amount.Sell, j.Amount.Buy, prov)
	default:
		return "", fmt.Errorf("未知的 Job 类型: %s", j.Type)
	}

	// 鲁棒性处理
	if len(j.Subscribers) == 0 {
		if constants.TGChatID != 0 {
			j.Subscribers = []int64{constants.TGChatID}
		}
	}

	executor.WithSubscribers(j.Subscribers)
	m.jobs.Store(j.String(), &Job{
		conf:     &j,
		executor: executor,
	})

	return j.String(), nil
}

func (m *Manager) DeleteJob(id string) error {
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

func (m *Manager) Unsubscribe(id string, chatID int64) error {
	if job, found := m.jobs.Load(id); found {
		subscribers := lo.Delete(job.conf.Subscribers, chatID)
		if len(subscribers) == 0 {
			return m.DeleteJob(id)
		}
		job.conf.Subscribers = subscribers
		job.executor.WithSubscribers(subscribers)
		return nil
	}
	return fmt.Errorf("job %s 不存在", id)
}

func (m *Manager) ClearAllJobs() {
	m.jobs.Range(func(key string, value *Job) bool {
		_ = m.DeleteJob(key)
		return true
	})
}

func (m *Manager) StartJob(id string) error {
	if job, found := m.jobs.Load(id); found {
		if job.cancel != nil {
			return fmt.Errorf("job %s 已在运行", id)
		}
		ctx, cancel := context.WithCancel(context.Background())
		job.cancel = cancel
		go func() {
			job.executor.Run(ctx)
			_ = m.StopJob(id)
		}()
		return nil
	}
	return fmt.Errorf("job %s 不存在", id)
}

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

func (m *Manager) ListAllJobs() []config.Job {
	var jobs []config.Job
	m.jobs.Range(func(key string, value *Job) bool {
		jobs = append(jobs, *value.conf)
		return true
	})
	return jobs
}

func (m *Manager) ListJobsBySubscriber(subId int64) []config.Job {
	var jobs []config.Job
	m.jobs.Range(func(key string, value *Job) bool {
		if slices.Contains(value.conf.Subscribers, subId) {
			jobs = append(jobs, *value.conf)
		}
		return true
	})
	return jobs
}

func (m *Manager) IsJobRunning(id string) bool {
	if job, found := m.jobs.Load(id); found {
		return job.cancel != nil
	}
	return false
}

func (m *Manager) GetJobData(id string) (config.Job, bool) {
	if job, found := m.jobs.Load(id); found {
		return *job.conf, true
	}
	return *new(config.Job), false
}
