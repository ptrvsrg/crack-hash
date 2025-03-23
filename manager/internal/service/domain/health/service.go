package health

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp"
	mongo2 "github.com/ptrvsrg/crack-hash/commonlib/storage/mongo"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
)

var (
	errAMQPReconnect = errors.New("amqp connection reconnect")
)

type svc struct {
	logger      zerolog.Logger
	mongoClient *mongo.Client
	amqpConn    *amqp.Connection
}

func NewService(logger zerolog.Logger, mongoClient *mongo.Client, amqpConn *amqp.Connection) domain.Health {
	return &svc{
		logger: logger.With().
			Str("type", "domain").
			Str("service", "health").
			Logger(),
		mongoClient: mongoClient,
		amqpConn:    amqpConn,
	}
}

func (s *svc) Health(ctx context.Context) error {
	s.logger.Info().Msg("health check")

	if err := mongo2.Ping(ctx, s.mongoClient); err != nil {
		s.logger.Error().Err(err).Msg("failed to check mongo client")
		return fmt.Errorf("failed to check mongo client: %w", err)
	}

	if s.amqpConn.IsReconnect() {
		s.logger.Error().Err(errAMQPReconnect).Msg("failed to check amqp connection")
		return fmt.Errorf("failed to check amqp connection: %w", errAMQPReconnect)
	}

	return nil
}
