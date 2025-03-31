package chunkbased

import (
	"crypto/md5" // nolint
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/ptrvsrg/crack-hash/worker/internal/combin"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
)

type svc struct {
	logger    zerolog.Logger
	chunkSize int
}

func NewService(logger zerolog.Logger, chunkSize int) infrastructure.HashBruteForce {
	return &svc{
		logger: logger.With().
			Str("type", "infrastructure").
			Str("service", "brute-force").
			Str("strategy", "chunk-based").Logger(),
		chunkSize: chunkSize,
	}
}

func (s *svc) BruteForceMD5(
	hash string, alphabet []string, maxLength, partNumber int, progressPeriod time.Duration,
) (<-chan infrastructure.TaskProgress, error) {

	s.logger.Info().
		Str("hash", hash).
		Int("maxLength", maxLength).
		Str("alphabet", strings.Join(alphabet, "")).
		Int("part", partNumber).
		Int("chunkSize", s.chunkSize).
		Msg("brute force md5")

	// Create alphabet iterator
	gen, err := combin.NewAlphabetIterator(
		strings.Join(alphabet, ""),
		maxLength,
		partNumber*s.chunkSize,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create alphabet iterator: %w", err)
	}

	// Initialize progress
	progress := infrastructure.TaskProgress{
		Answers: make([]string, 0, 1024),
		Percent: 0.0,
		Status:  infrastructure.TaskStatusInProgress,
	}
	progressCh := make(chan infrastructure.TaskProgress, 1)

	go func() {
		// Create ticker
		ticker := time.NewTicker(progressPeriod)
		defer ticker.Stop()
		defer close(progressCh)

		for i := 0; i < s.chunkSize && gen.Next(); i++ {
			select {
			case <-ticker.C:
				progress.Percent = 100 * float64(i+1) / float64(s.chunkSize)
				s.logger.Debug().
					Str("hash", hash).
					Int("maxLength", maxLength).
					Int("part", partNumber).
					Msgf("processed by %.2f%%", progress.Percent)

				progressCh <- progress

			default:
				word := gen.Current()
				md5Hash := md5.Sum([]byte(word)) // nolint
				sum := hex.EncodeToString(md5Hash[:])

				if sum == hash {
					progress.Answers = append(progress.Answers, word)
				}

			}
		}

		progress.Percent = 100.0
		progress.Status = infrastructure.TaskStatusSuccess

		s.logger.Debug().
			Str("hash", hash).
			Int("maxLength", maxLength).
			Int("part", partNumber).
			Msgf("processed by %.2f%%", progress.Percent)

		progressCh <- progress
	}()

	return progressCh, nil
}
