package hashcrack

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron"

	"github.com/ptrvsrg/crack-hash/commonlib/cron"
	"github.com/ptrvsrg/crack-hash/manager/internal/di"
)

func RegisterDeleteExpiredTaskJob(c *di.Container) cron.RegisterFunc {
	return func(ctx context.Context, scheduler *gocron.Scheduler) error {
		logger := c.Logger.With().
			Str("component", "cron-scheduler").
			Str("job", "delete-expired-task").
			Logger()

		_, err := scheduler.
			Every(c.Config.Task.MaxAge/2).
			Do(
				func(ctx context.Context) {
					logger.Debug().Msg("running cron job")

					if err := c.DomainSVCs.HashCrackTask.DeleteExpiredTasks(ctx); err != nil {
						logger.Error().Err(err).Stack().Msg("failed to delete expired tasks")
					}
				}, ctx,
			)

		if err != nil {
			return fmt.Errorf("failed to register cron job: %w", err)
		}

		return nil
	}
}
