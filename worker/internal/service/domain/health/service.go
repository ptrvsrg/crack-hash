package health

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
)

var (
	errAMQPReconnect = errors.New("amqp connection reconnect")
)

type svc struct {
	logger   zerolog.Logger
	amqpConn *amqp.Connection
}

func NewService(logger zerolog.Logger, amqpConn *amqp.Connection) domain.Health {
	return &svc{
		logger: logger.With().
			Str("type", "domain").
			Str("service", "health").
			Logger(),
		amqpConn: amqpConn,
	}
}

func (s *svc) Health(_ context.Context) error {
	s.logger.Info().Msg("health check")

	if s.amqpConn.IsReconnect() {
		s.logger.Error().Err(errAMQPReconnect).Msg("failed to check amqp connection")
		return errAMQPReconnect
	}

	return nil
}
