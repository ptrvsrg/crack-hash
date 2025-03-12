package chunkbased

import (
	"crypto/md5" // nolint
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ptrvsrg/crack-hash/worker/internal/combin"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
)

type svc struct {
	logger    zerolog.Logger
	chunkSize int
}

func NewService(chunkSize int) infrastructure.HashBruteForce {
	return &svc{
		logger: log.With().
			Str("type", "infrastructure").
			Str("service", "brute-force").
			Str("strategy", "chunk-based").Logger(),
		chunkSize: chunkSize,
	}
}

func (s *svc) BruteForceMD5(hash string, alphabet []string, maxLength, partNumber int) ([]string, error) {
	s.logger.Info().
		Str("hash", hash).
		Int("maxLength", maxLength).
		Str("alphabet", strings.Join(alphabet, "")).
		Int("part", partNumber).
		Int("chunkSize", s.chunkSize).
		Msg("brute force md5")

	gen, err := combin.NewAlphabetIterator(
		strings.Join(alphabet, ""),
		maxLength,
		partNumber*s.chunkSize,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create alphabet iterator: %w", err)
	}

	results := make([]string, 0, 1024)
	for i := 0; i < s.chunkSize && gen.Next(); i++ {
		word := gen.Current()
		md5Hash := md5.Sum([]byte(word)) // nolint
		sum := hex.EncodeToString(md5Hash[:])

		if sum == hash {
			results = append(results, word)
		}

		if (i+1)%100_000 == 0 {
			percent := 100 * float64(i+1) / float64(s.chunkSize)
			s.logger.Debug().
				Str("hash", hash).
				Int("maxLength", maxLength).
				Int("part", partNumber).
				Msgf("processed by %.2f%%", percent)
		}
	}

	return results, nil
}
