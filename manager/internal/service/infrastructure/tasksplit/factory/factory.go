package factory

import (
	"github.com/rs/zerolog"

	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/tasksplit/chunkbased"
)

type Strategy string

const (
	StrategyChunkBased Strategy = "chunk-based"
)

func NewService(logger zerolog.Logger, cfg config.TaskSplitConfig) infrastructure.TaskSplit {
	switch Strategy(cfg.Strategy) {
	case StrategyChunkBased:
		fallthrough
	default:
		return chunkbased.NewService(logger, cfg.ChunkSize)
	}
}
