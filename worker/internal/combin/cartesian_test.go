package combin_test

import (
	"math"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/worker/internal/combin"
)

func init() {
	logging.Setup(true)
}

func TestAlphabetIterator_Next(t *testing.T) {
	alphabet := "abc"
	maxLength := 3

	t.Run(
		"Without start index", func(t *testing.T) {
			iterator, err := combin.NewAlphabetIterator(alphabet, maxLength, 0)
			require.NoError(t, err)

			results := make([]string, 0)
			i := 0
			for iterator.Next() {
				results = append(results, iterator.Current())
				i++
			}

			log.Info().Msgf("results: %v", results)
			require.Equal(t, int(sumOfGeomSeries(float64(len(alphabet)), float64(len(alphabet)), maxLength)), i)
		},
	)

	t.Run(
		"With start index", func(t *testing.T) {
			startIndex := 10

			iterator, err := combin.NewAlphabetIterator(alphabet, maxLength, startIndex)
			require.NoError(t, err)

			results := make([]string, 0)
			i := 0
			for iterator.Next() {
				results = append(results, iterator.Current())
				i++
			}

			log.Info().Msgf("results: %v", results)
			require.Equal(
				t, int(sumOfGeomSeries(float64(len(alphabet)), float64(len(alphabet)), maxLength))-startIndex, i,
			)
		},
	)
}

func sumOfGeomSeries(a, r float64, n int) float64 {
	return a * (math.Pow(r, float64(n)) - 1) / (r - 1)
}
