package client

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"resty.dev/v3"
	"time"
)

type logger struct {
	logger zerolog.Logger
}

func (l *logger) Errorf(format string, v ...any) {
	l.logger.Error().Msgf(format, v...)
}

func (l *logger) Warnf(format string, v ...any) {
	l.logger.Warn().Msgf(format, v...)
}

func (l *logger) Debugf(format string, v ...any) {
	l.logger.Debug().Msgf(format, v...)
}

func New() *resty.Client {
	cb := resty.NewCircuitBreaker().
		SetTimeout(5 * time.Second).
		SetFailureThreshold(10).
		SetSuccessThreshold(5)

	l := &logger{
		logger: log.With().Str("component", "http-client").Logger(),
	}

	client := resty.New().
		SetCircuitBreaker(cb).
		SetRetryCount(3).
		EnableDebug().
		EnableTrace().
		SetLogger(l)

	return client
}
