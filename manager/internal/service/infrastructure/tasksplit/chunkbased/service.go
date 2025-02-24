package chunkbased

import (
	"context"
	"fmt"
	"github.com/ptrvsrg/crack-hash/manager/internal/helper"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
)

var (
	maxWordsPerSubtask = 10_000_000
)

type svc struct {
	logger zerolog.Logger
}

func NewService() infrastructure.TaskSplit {
	return &svc{
		logger: log.With().Str("infra-service", "task-split").Logger(),
	}
}

func (s *svc) Split(_ context.Context, wordMaxLength, alphabetLength int) (int, error) {
	s.logger.Info().Msgf("split task: wordMaxLength=%d, alphabetLength=%d", wordMaxLength, alphabetLength)

	// Validate input
	if wordMaxLength <= -1 {
		return 0, infrastructure.ErrInvalidWordMaxLength
	}

	if alphabetLength <= -1 {
		return 0, infrastructure.ErrInvalidAlphabetLength
	}

	// Calculate word count
	wordCount, err := helper.SumOfGeomSeries(alphabetLength, alphabetLength, wordMaxLength)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to calculate number of subtasks")
		return 0, fmt.Errorf("failed to calculate number of subtasks: %w", err)
	}

	// Calculate number of subtasks
	numSubtasks := math.Ceil(float64(wordCount) / float64(maxWordsPerSubtask))

	return int(numSubtasks), nil
}
