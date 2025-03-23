package di

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"resty.dev/v3"

	"github.com/ptrvsrg/crack-hash/commonlib/http/client"
	"github.com/ptrvsrg/crack-hash/commonlib/http/client/loadbalancer"
	"github.com/ptrvsrg/crack-hash/commonlib/http/handler"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository/memory/hashcracktask"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/tasksplit/factory"
	hashcrackhdlr "github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/health"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/swagger"
)

type Container struct {
	Config     config.Config
	Logger     zerolog.Logger
	HTTPClient *resty.Client
	Repos      repository.Repositories
	InfraSVCs  infrastructure.Services
	DomainSVCs domain.Services
	Handlers   []handler.Handler
}

func NewContainer(cfg config.Config) *Container {
	c := &Container{
		Config: cfg,
		Logger: log.Logger,
	}

	c.setupHTTPClient()
	c.setupRepositories()
	c.setupServices()
	c.setupHandlers()

	return c
}

func (c *Container) setupHTTPClient() {
	c.Logger.Info().Msg("setup HTTP client")

	var err error
	c.HTTPClient, err = client.New(
		client.WithLoadBalancer(
			c.Config.Worker.Addresses,
			loadbalancer.WithHealthChecks(
				c.Config.Worker.Health.Path,
				c.Config.Worker.Health.Timeout,
				c.Config.Worker.Health.Interval,
				c.Config.Worker.Health.Retries,
			),
		),
		client.WithRetries(3, 5*time.Second, 10*time.Second),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create http client")
	}
}

func (c *Container) setupRepositories() {
	c.Logger.Info().Msg("setup repositories")

	c.Repos = repository.Repositories{
		HashCrackTask: hashcracktask.NewRepo(),
	}
}

func (c *Container) setupServices() {
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
			c.HTTPClient,
			c.Repos.HashCrackTask,
			c.InfraSVCs.TaskSplit,
		),
	}
}

func (c *Container) setupHandlers() {
	c.Logger.Info().Msg("setup handlers")

	c.Handlers = []handler.Handler{
		health.NewHandler(),
		swagger.NewHandler(),
		hashcrackhdlr.NewHandler(c.DomainSVCs.HashCrackTask),
	}
}

func (c *Container) Close() error {
	c.Logger.Info().Msg("closing container")

	errs := make([]error, 0)

	c.Logger.Info().Msg("closing HTTP client")
	if err := c.HTTPClient.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close container: %w", errors.Join(errs...))
	}

	return nil
}
