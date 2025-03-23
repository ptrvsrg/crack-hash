package chunkbased

import (
	"context"
	"fmt"
	"math"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ptrvsrg/crack-hash/manager/internal/helper"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
)

type svc struct {
	chunkSize int
	logger    zerolog.Logger
}

func NewService(chunkSize int) infrastructure.TaskSplit {
	return &svc{
		chunkSize: chunkSize,
		logger: log.With().
			Str("type", "infrastructure").
			Str("infra-service", "task-split").
			Logger(),
	}
}

func (s *svc) Split(_ context.Context, wordMaxLength, alphabetLength int) (int, error) {
	s.logger.Info().
		Int("wordMaxLength", wordMaxLength).
		Int("alphabetLength", alphabetLength).
		Msg("split task")

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
	numSubtasks := int(math.Ceil(float64(wordCount) / float64(s.chunkSize)))

	s.logger.Info().
		Int("numSubtasks", numSubtasks).
		Msg("number of subtasks calculated")

	return numSubtasks, nil
}
