package factory

import (
	"errors"
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

func NewService(strategy Strategy) (infrastructure.HashBruteForce, error) {
	switch strategy {
	case StrategyChunkBased:
		return chunkbased.NewService(), nil
	default:
		return nil, ErrInvalidStrategy
	}
}
