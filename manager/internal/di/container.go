package di

import (
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/client"
	"github.com/ptrvsrg/crack-hash/manager/internal/cron/executor"
	hashcrack2 "github.com/ptrvsrg/crack-hash/manager/internal/cron/executor/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository/memory"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/tasksplit/factory"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler"
	hashcrack1 "github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/health"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler/swagger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Container struct {
	Config     config.Config
	Logger     zerolog.Logger
	Repos      repository.Repositories
	InfraSVCs  infrastructure.Services
	DomainSVCs domain.Services
	Executors  []executor.Executor
	Handlers   []handler.Handler
}

func NewContainer(cfg config.Config) *Container {
	c := &Container{
		Config: cfg,
		Logger: log.Logger,
	}

	c.Logger.Info().Msg("setup repositories")
	c.Repos = repository.Repositories{
		HashCrackTask: memory.NewRepo(),
	}

	c.Logger.Info().Msg("setup services")

	splitSvc, err := factory.NewService(c.Config.Task.SplitStrategy)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create task split strategy")
	}

	c.InfraSVCs = infrastructure.Services{
		TaskSplit: splitSvc,
	}
	c.DomainSVCs = domain.Services{
		HashCrackTask: hashcrack.NewService(c.Config.Worker, client.New(), c.Repos.HashCrackTask, c.InfraSVCs.TaskSplit),
	}

	c.Logger.Info().Msg("setup handlers")
	c.Handlers = []handler.Handler{
		health.NewHandler(),
		swagger.NewHandler(),
		hashcrack1.NewHandler(c.DomainSVCs.HashCrackTask),
	}

	c.Logger.Info().Msg("setup executors")
	c.Executors = []executor.Executor{
		hashcrack2.NewExecutor(c.DomainSVCs.HashCrackTask),
	}

	return c
}

func (c *Container) Close() error {
	return nil
}
