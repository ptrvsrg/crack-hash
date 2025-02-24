package chunkbased

import (
	"crypto/md5" //nolint
	"encoding/hex"
	"fmt"
	"github.com/ptrvsrg/crack-hash/worker/internal/combin"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"hash"
	"strings"
)

var (
	maxWordsPerSubtask = 10_000_000
)

type svc struct {
	logger zerolog.Logger
	digest hash.Hash
}

func NewService() infrastructure.HashBruteForce {
	return &svc{
		logger: log.With().
			Str("service", "brute-force").
			Str("strategy", "chunk-based").Logger(),
		digest: md5.New(), //nolint
	}
}

func (s svc) BruteForceMD5(hash string, alphabet []string, maxLength, partNumber int) ([]string, error) {
	s.logger.Info().Msg("brute force md5")

	gen, err := combin.NewAlphabetIterator(
		strings.Join(alphabet, ""),
		maxLength,
		partNumber*maxWordsPerSubtask,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create alphabet iterator: %w", err)
	}

	resultCh := make(chan string, 1024)
	for i := 0; i < maxWordsPerSubtask && gen.Next(); i++ {
		word := gen.Current()
		md5Hash := s.digest.Sum([]byte(word))
		sum := hex.EncodeToString(md5Hash[:])

		if sum == hash {
			resultCh <- word
		}

		if (i+1)%1_000_000 == 0 {
			s.logger.Debug().Msgf("processed %d words", i+1)
		}
	}

	close(resultCh)

	results := make([]string, 0)
	for result := range resultCh {
		results = append(results, result)
	}

	return results, nil
}
