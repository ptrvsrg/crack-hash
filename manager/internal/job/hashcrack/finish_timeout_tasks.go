package hashcrack

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron"

	"github.com/ptrvsrg/crack-hash/commonlib/cron"
	"github.com/ptrvsrg/crack-hash/manager/internal/di"
)

func RegisterFinishTimeoutTasksJob(c *di.Container) cron.RegisterFunc {
	return func(ctx context.Context, scheduler *gocron.Scheduler) error {
		logger := c.Logger.With().
			Str("component", "cron-scheduler").
			Str("job", "finish-timeout-task").
			Logger()

		_, err := scheduler.
			Every(c.Config.Task.FinishDelay).
			Do(
				func(ctx context.Context) {
					logger.Debug().Msg("running cron job")

					if err := c.DomainSVCs.HashCrackTask.FinishTimeoutTasks(ctx); err != nil {
						logger.Error().Err(err).Stack().Msg("failed to finish timeout tasks")
					}
				}, ctx,
			)

		if err != nil {
			return fmt.Errorf("failed to register cron job: %w", err)
		}

		return nil
	}
}
