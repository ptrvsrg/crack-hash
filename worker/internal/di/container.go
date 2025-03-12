package di

import (
	"errors"
	"fmt"
	"time"

	jobqueue "github.com/dirkaholic/kyoo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"resty.dev/v3"

	"github.com/ptrvsrg/crack-hash/commonlib/http/client"
	"github.com/ptrvsrg/crack-hash/commonlib/http/handler"
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain/hashcracktask"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/bruteforce/factory"
	hashcrackhdlr "github.com/ptrvsrg/crack-hash/worker/internal/transport/http/handler/hashcracktask"
	"github.com/ptrvsrg/crack-hash/worker/internal/transport/http/handler/health"
	"github.com/ptrvsrg/crack-hash/worker/internal/transport/http/handler/swagger"
)

type Container struct {
	Config     config.Config
	Logger     zerolog.Logger
	HTTPClient *resty.Client
	JobQueue   *jobqueue.JobQueue
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
	c.setupJobQueue()
	c.setupServices()
	c.setupHandlers()

	return c
}

func (c *Container) Close() error {
	c.Logger.Info().Msg("closing container")

	errs := make([]error, 0)

	c.Logger.Info().Msg("closing HTTP client")
	if err := c.HTTPClient.Close(); err != nil {
		errs = append(errs, err)
	}

	c.Logger.Info().Msg("closing job queue")
	c.JobQueue.Stop()

	if len(errs) > 0 {
		return fmt.Errorf("failed to close container: %w", errors.Join(errs...))
	}

	return nil
}

func (c *Container) setupHTTPClient() {
	c.Logger.Info().Msg("setup HTTP client")

	var err error
	c.HTTPClient, err = client.New(
		client.WithRetries(3, 5*time.Second, 10*time.Second),
	)
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to create HTTP client")
	}
}

func (c *Container) setupJobQueue() {
	c.Logger.Info().Msg("setup job queue")

	c.JobQueue = jobqueue.NewJobQueue(c.Config.Task.Concurrency)
	c.JobQueue.Start()
}

func (c *Container) setupServices() {
	c.Logger.Info().Msg("setup services")

	bruteForceSvc, err := factory.NewService(c.Config.Task.Split)
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to create brute force service")
	}

	c.InfraSVCs = infrastructure.Services{
		HashBruteForce: bruteForceSvc,
	}
	c.DomainSVCs = domain.Services{
		HashCrackTask: hashcracktask.NewService(
			c.Config.Manager,
			c.HTTPClient,
			c.JobQueue,
			c.InfraSVCs.HashBruteForce,
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
