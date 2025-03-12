package cron

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

type RegisterFunc func(ctx context.Context, scheduler *gocron.Scheduler) error

func NewScheduler(ctx context.Context, registerFuncs ...RegisterFunc) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.UTC)

	errs := make([]error, 0)
	for _, rf := range registerFuncs {
		if err := rf(ctx, scheduler); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		err := fmt.Errorf("failed to register cron jobs: %w", errors.Join(errs...)) // nolint
		log.Fatal().Err(err).Msg("failed to register cron jobs")
	}

	return scheduler
}
