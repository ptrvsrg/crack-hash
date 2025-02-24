package di

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp"
	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/consumer"
	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher"
	"github.com/ptrvsrg/crack-hash/commonlib/http/handler"
	mongo2 "github.com/ptrvsrg/crack-hash/commonlib/storage/mongo"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/bus/amqp/consumer/taskresult"
	publisher2 "github.com/ptrvsrg/crack-hash/manager/internal/bus/amqp/publisher"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository/mongo/hashcracktask"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/tasksplit/factory"
	hashcrackhdlr "github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/health"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/swagger"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
)

type Providers struct {
	AMQPConn    *amqp.Connection
	AMQPChannel *amqp.Channel
	MongoDB     *mongo.Client
}

type Container struct {
	Config     config.Config
	Logger     zerolog.Logger
	Providers  Providers
	Publishers publisher2.Publishers
	Repos      repository.Repositories
	InfraSVCs  infrastructure.Services
	DomainSVCs domain.Services
	Handlers   []handler.Handler
	Consumers  []consumer.Consumer
}

func NewContainer(ctx context.Context, cfg config.Config) *Container {
	c := &Container{
		Config: cfg,
		Logger: log.Logger,
	}

	c.setupProviders(ctx)
	c.setupRepositories(ctx)
	c.setupPublishers(ctx)
	c.setupServices(ctx)
	c.setupHandlers(ctx)
	c.setupConsumers(ctx)

	return c
}

func (c *Container) Close(ctx context.Context) error {
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

	c.Logger.Info().Msg("closing MongoDB")

	ctx, cansel := context.WithTimeout(ctx, time.Second*10)
	defer cansel()

	if err := c.Providers.MongoDB.Disconnect(ctx); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close container: %w", errors.Join(errs...))
	}

	return nil
}

func (c *Container) setupProviders(ctx context.Context) {
	c.Logger.Info().Msg("setup MongoDB client")
	mongoClient, err := mongo2.NewClient(
		ctx,
		mongo2.Config{
			URI:      c.Config.MongoDB.URI,
			Username: c.Config.MongoDB.Username,
			Password: c.Config.MongoDB.Password,
		},
	)
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to setup MongoDB client")
	}

	c.Logger.Info().Msg("setup AMQP connection")

	var (
		amqpConn *amqp.Connection
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
		MongoDB:     mongoClient,
	}
}

func (c *Container) setupRepositories(_ context.Context) {
	c.Logger.Info().Msg("setup repositories")

	c.Repos = repository.Repositories{
		HashCrackTask: hashcracktask.NewRepo(
			c.Providers.MongoDB, c.Config.MongoDB,
		),
	}
}

func (c *Container) setupPublishers(_ context.Context) {
	c.Logger.Info().Msg("setup publishers")

	c.Publishers = publisher2.Publishers{
		TaskStarted: publisher.New[message.HashCrackTaskStarted](
			c.Providers.AMQPChannel,
			publisher.Config{
				Exchange:   c.Config.AMQP.Publishers.TaskStarted.Exchange,
				RoutingKey: c.Config.AMQP.Publishers.TaskStarted.RoutingKey,
			},
		),
	}
}

func (c *Container) setupServices(_ context.Context) {
	c.Logger.Info().Msg("setup services")

	splitSvc, err := factory.NewService(c.Config.Task.Split)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create task split strategy")
	}

	c.InfraSVCs = infrastructure.Services{
		TaskSplit: splitSvc,
	}
	c.DomainSVCs = domain.Services{
		HashCrackTask: hashcrack.NewService(
			c.Config.Task,
			c.Repos.HashCrackTask,
			c.InfraSVCs.TaskSplit,
			c.Publishers.TaskStarted,
		),
	}
}

func (c *Container) setupHandlers(_ context.Context) {
	c.Logger.Info().Msg("setup handlers")

	c.Handlers = []handler.Handler{
		health.NewHandler(),
		swagger.NewHandler(),
		hashcrackhdlr.NewHandler(c.DomainSVCs.HashCrackTask),
	}
}

func (c *Container) setupConsumers(_ context.Context) {
	c.Logger.Info().Msg("setup consumers")

	c.Consumers = []consumer.Consumer{
		taskresult.NewConsumer(
			c.Providers.AMQPChannel, c.Config.AMQP.Consumers.TaskResult, c.DomainSVCs.HashCrackTask,
		),
	}
}
