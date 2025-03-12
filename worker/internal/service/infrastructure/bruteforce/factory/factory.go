package factory

import (
	"errors"

	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/bruteforce/chunkbased"
)

var (
	ErrInvalidStrategy = errors.New("invalid strategy")
)

type Strategy string

const (
	StrategyChunkBased Strategy = "chunk-based"
)

func NewService(cfg config.TaskSplitConfig) (infrastructure.HashBruteForce, error) {
	switch Strategy(cfg.Strategy) {

	case StrategyChunkBased:
		return chunkbased.NewService(cfg.ChunkSize), nil

	default:
		return nil, ErrInvalidStrategy
	}
}
