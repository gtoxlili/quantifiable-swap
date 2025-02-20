package job

import (
	"github.com/gtoxlili/quantifiable-swap/common/config"
)

type IManager interface {
	CreateJob(j config.Job) (string, error)
	DeleteJob(id string) error
	Unsubscribe(id string, chatID int64) error
	ClearAllJobs()
	StartJob(id string) error
	StopJob(id string) error
	ListJobsBySubscriber(subId int64) []config.Job
	ListAllJobs() []config.Job
	IsJobRunning(id string) bool
	GetJobData(id string) (config.Job, bool)
}
