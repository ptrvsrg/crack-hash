package cron

import (
	"context"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
)

type Job func(context.Context)

func NewSchedulerOrDie() gocron.Scheduler {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal().Err(err).Stack().Msgf("failed to create scheduler")
	}
	return scheduler
}
