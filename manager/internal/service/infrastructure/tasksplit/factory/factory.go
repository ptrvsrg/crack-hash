package factory

import (
	"errors"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/tasksplit/chunkbased"
)

var (
	ErrInvalidStrategy = errors.New("invalid strategy")
)

type Strategy string

const (
	StrategyChunkBased Strategy = "chunk-based"
)

func NewService(strategy Strategy) (infrastructure.TaskSplit, error) {
	switch strategy {
	case StrategyChunkBased:
		return chunkbased.NewService(), nil
	default:
		return nil, ErrInvalidStrategy
	}
}
