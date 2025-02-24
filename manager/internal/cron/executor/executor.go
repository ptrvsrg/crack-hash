package executor

import (
	"github.com/go-co-op/gocron/v2"
)

type RegisterFunc func(scheduler gocron.Scheduler) error

type Executor interface {
	RegisterJobs(scheduler gocron.Scheduler) error
}
