package factory

import (
	"github.com/rs/zerolog"

	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/bruteforce/chunkbased"
)

type Strategy string

const (
	StrategyChunkBased Strategy = "chunk-based"
)

func NewService(logger zerolog.Logger, cfg config.TaskSplitConfig) infrastructure.HashBruteForce {
	switch Strategy(cfg.Strategy) {
	case StrategyChunkBased:
		fallthrough
	default:
		return chunkbased.NewService(logger, cfg.ChunkSize)
	}
}
