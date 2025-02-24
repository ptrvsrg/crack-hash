package hashcrack

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/ptrvsrg/crack-hash/manager/internal/cron/executor"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type exec struct {
	svc    domain.HashCrackTask
	logger zerolog.Logger
}

func NewExecutor(svc domain.HashCrackTask) executor.Executor {
	return &exec{
		svc:    svc,
		logger: log.With().Str("executor", "hash-crack").Logger(),
	}
}

func (e *exec) RegisterJobs(scheduler gocron.Scheduler) error {
	e.logger.Debug().Msg("register jobs")

	registerAll := func(registerFuncs ...executor.RegisterFunc) error {
		for _, registerFunc := range registerFuncs {
			if err := registerFunc(scheduler); err != nil {
				return err
			}
		}

		return nil
	}

	err := registerAll(
		e.registerFinishTasksJob,
	)
	if err != nil {
		e.logger.Error().Err(err).Stack().Msg("failed to register jobs")
		return fmt.Errorf("failed to register jobs: %w", err)
	}

	return nil
}

func (e *exec) registerFinishTasksJob(scheduler gocron.Scheduler) error {
	e.logger.Debug().Msg("register finish tasks job")

	_, err := scheduler.NewJob(
		gocron.CronJob("* * * * *", false),
		gocron.NewTask(e.finishTasks),
	)

	if err != nil {
		return fmt.Errorf("failed to register finish tasks job: %w", err)
	}

	return nil
}

func (e *exec) finishTasks() {
	e.logger.Debug().Msg("finish crack hash task")
	_ = e.svc.FinishTasks(context.Background())
}
