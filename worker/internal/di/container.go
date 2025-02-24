package di

import (
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/client"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain/hashcracktask"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/bruteforce/factory"
	"github.com/ptrvsrg/crack-hash/worker/internal/transport/http/handler"
	hashcrack2 "github.com/ptrvsrg/crack-hash/worker/internal/transport/http/handler/hashcracktask"
	"github.com/ptrvsrg/crack-hash/worker/internal/transport/http/handler/health"
	"github.com/ptrvsrg/crack-hash/worker/internal/transport/http/handler/swagger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Container struct {
	Config     config.Config
	Logger     zerolog.Logger
	InfraSVCs  infrastructure.Services
	DomainSVCs domain.Services
	Handlers   []handler.Handler
}

func NewContainer(cfg config.Config) *Container {
	c := &Container{
		Config: cfg,
		Logger: log.Logger,
	}

	c.Logger.Info().Msg("setup services")

	bruteForceSvc, err := factory.NewService(cfg.Task.SplitStrategy)
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to create brute force service")
	}

	c.InfraSVCs = infrastructure.Services{
		HashBruteForce: bruteForceSvc,
	}
	c.DomainSVCs = domain.Services{
		HashCrackTask: hashcracktask.NewService(cfg.Manager, client.New(), c.InfraSVCs.HashBruteForce),
	}

	c.Logger.Info().Msg("setup handlers")
	c.Handlers = []handler.Handler{
		health.NewHandler(),
		swagger.NewHandler(),
		hashcrack2.NewHandler(c.DomainSVCs.HashCrackTask),
	}

	return c
}

func (c *Container) Close() error {
	return nil
}
