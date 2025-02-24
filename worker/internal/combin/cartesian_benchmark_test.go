package combin

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func BenchmarkAlphabetIterator_Next(b *testing.B) {
	alphabet := "abcdefghijklmnopqrstuvwxyz0123456789"
	maxLength := 100

	iterator, err := NewAlphabetIterator(alphabet, maxLength, 0)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		iterator.Next()
	}
}
