package di

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp"
	consumer2 "github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/consumer"
	publisher2 "github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher"
	"github.com/ptrvsrg/crack-hash/commonlib/http/handler"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/bus/amqp/consumer/taskstarted"
	"github.com/ptrvsrg/crack-hash/worker/internal/bus/amqp/publisher"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain/hashcracktask"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain/health"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/bruteforce/factory"
	healthhdlr "github.com/ptrvsrg/crack-hash/worker/internal/transport/http/handler/health"
)

type Providers struct {
	AMQPConn    *amqp.Connection
	AMQPChannel *amqp.Channel
}

type Container struct {
	Config     config.Config
	Logger     zerolog.Logger
	Providers  Providers
	Publishers publisher.Publishers
	InfraSVCs  infrastructure.Services
	DomainSVCs domain.Services
	Handlers   []handler.Handler
	Consumers  []consumer2.Consumer
}

func NewContainer(ctx context.Context, cfg config.Config) *Container {
	c := &Container{
		Config: cfg,
		Logger: log.Logger,
	}

	c.setupProviders(ctx)
	c.setupPublishers(ctx)
	c.setupServices(ctx)
	c.setupHandlers(ctx)
	c.setupConsumers(ctx)

	return c
}

func (c *Container) Close(_ context.Context) error {
	c.Logger.Info().Msg("closing container")

	errs := make([]error, 0)

	c.Logger.Info().Msg("closing AMQP channel")
	if err := c.Providers.AMQPChannel.Close(); err != nil {
		errs = append(errs, err)
	}

	c.Logger.Info().Msg("closing AMQP connection")
	if err := c.Providers.AMQPConn.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close container: %w", errors.Join(errs...))
	}

	return nil
}

func (c *Container) setupProviders(ctx context.Context) {
	c.Logger.Info().Msg("setup AMQP connection")
	var (
		amqpConn *amqp.Connection
		err      error
	)
	if len(c.Config.AMQP.URIs) == 1 {
		amqpConn, err = amqp.Dial(
			ctx,
			amqp.Config{
				URI:      c.Config.AMQP.URIs[0],
				Username: c.Config.AMQP.Username,
				Password: c.Config.AMQP.Password,
				Prefetch: c.Config.AMQP.Prefetch,
			},
		)
	} else {
		amqpConn, err = amqp.DialCluster(
			ctx,
			amqp.ClusterConfig{
				URIs:     c.Config.AMQP.URIs,
				Username: c.Config.AMQP.Username,
				Password: c.Config.AMQP.Password,
				Prefetch: c.Config.AMQP.Prefetch,
			},
		)
	}
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to setup AMQP connection")
	}

	c.Logger.Info().Msg("setup AMQP channel")
	amqpCh, err := amqpConn.Channel(ctx)
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to setup AMQP channel")
	}

	c.Providers = Providers{
		AMQPConn:    amqpConn,
		AMQPChannel: amqpCh,
	}
}

func (c *Container) setupPublishers(_ context.Context) {
	c.Logger.Info().Msg("setup publishers")

	c.Publishers = publisher.Publishers{
		TaskResult: publisher2.New[message.HashCrackTaskResult](
			c.Providers.AMQPChannel,
			publisher2.Config{
				Exchange:   c.Config.AMQP.Publishers.TaskResult.Exchange,
				RoutingKey: c.Config.AMQP.Publishers.TaskResult.RoutingKey,
			},
		),
	}
}

func (c *Container) setupServices(_ context.Context) {
	c.Logger.Info().Msg("setup services")

	c.InfraSVCs = infrastructure.Services{
		HashBruteForce: factory.NewService(c.Logger, c.Config.Task.Split),
	}
	c.DomainSVCs = domain.Services{
		HashCrackTask: hashcracktask.NewService(
			c.Logger,
			c.Config.Task.ProgressPeriod,
			c.Publishers.TaskResult,
			c.InfraSVCs.HashBruteForce,
		),
		Health: health.NewService(c.Logger, c.Providers.AMQPConn),
	}
}

func (c *Container) setupHandlers(_ context.Context) {
	c.Logger.Info().Msg("setup handlers")

	c.Handlers = []handler.Handler{
		healthhdlr.NewHandler(c.Logger, c.DomainSVCs.Health),
	}
}

func (c *Container) setupConsumers(_ context.Context) {
	c.Logger.Info().Msg("setup consumers")

	c.Consumers = []consumer2.Consumer{
		taskstarted.NewConsumer(
			c.Providers.AMQPChannel, c.Config.AMQP.Consumers.TaskStarted, c.DomainSVCs.HashCrackTask,
		),
	}
}
