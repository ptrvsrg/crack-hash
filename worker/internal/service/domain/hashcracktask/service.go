package hashcracktask

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
)

type svc struct {
	logger         zerolog.Logger
	progressPeriod time.Duration
	publisher      publisher.Publisher[message.HashCrackTaskResult]
	bruteforce     infrastructure.HashBruteForce
}

func NewService(
	logger zerolog.Logger,
	progressPeriod time.Duration,
	publisher publisher.Publisher[message.HashCrackTaskResult],
	bruteforce infrastructure.HashBruteForce,
) domain.HashCrackTask {
	return &svc{
		logger: logger.With().
			Str("type", "domain").
			Str("service", "hash-crack-task").
			Logger(),
		progressPeriod: progressPeriod,
		publisher:      publisher,
		bruteforce:     bruteforce,
	}
}

func (s *svc) ExecuteTask(ctx context.Context, input *message.HashCrackTaskStarted) error {
	s.logger.Info().
		Str("id", input.RequestID).
		Int("part", input.PartNumber).
		Msg("start brute force md5")

	go s.executeTask(ctx, input)

	return nil
}

func (s *svc) executeTask(ctx context.Context, input *message.HashCrackTaskStarted) {
	// Brute force
	s.logger.Info().
		Str("id", input.RequestID).
		Int("part", input.PartNumber).
		Msg("brute force md5")

	progressCh, err := s.bruteforce.BruteForceMD5(
		input.Hash, input.Alphabet.Symbols, input.MaxLength, input.PartNumber, s.progressPeriod,
	)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to brute force md5")

		msg := buildErrorResultMessage(input.RequestID, input.PartNumber, lo.ToPtr(err.Error()))
		if err := s.publisher.SendMessage(ctx, msg, publisher.Persistent, false, false); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to send result message")
		}

		return
	}

	for progress := range progressCh {
		// Send result
		var msg *message.HashCrackTaskResult
		if progress.Status == infrastructure.TaskStatusError {
			msg = buildErrorResultMessage(input.RequestID, input.PartNumber, progress.Reason)
		} else {
			msg = buildResultMessage(input.RequestID, input.PartNumber, progress)
		}

		if err := s.publisher.SendMessage(ctx, msg, publisher.Persistent, false, false); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to send result message")
			return
		}
	}

	s.logger.Info().
		Str("id", input.RequestID).
		Int("part", input.PartNumber).
		Msg("end brute force md5")
}

func buildErrorResultMessage(requestID string, partNumber int, error *string) *message.HashCrackTaskResult {
	return &message.HashCrackTaskResult{
		RequestID:  requestID,
		PartNumber: partNumber,
		Error:      error,
		Status:     string(infrastructure.TaskStatusError),
	}
}

func buildResultMessage(
	requestID string, partNumber int, progress infrastructure.TaskProgress,
) *message.HashCrackTaskResult {
	return &message.HashCrackTaskResult{
		RequestID:  requestID,
		PartNumber: partNumber,
		Status:     string(progress.Status),
		Answer: &message.Answer{
			Words:   progress.Answers,
			Percent: progress.Percent,
		},
	}
}
