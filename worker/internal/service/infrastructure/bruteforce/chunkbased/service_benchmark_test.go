package chunkbased_test

import (
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/bruteforce/chunkbased"
	"github.com/rs/zerolog"
	"strings"
	"testing"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func Benchmark(b *testing.B) {
	svc := chunkbased.NewService()
	hash := "abcde"
	alphabet := "abcdefghijklmnopqrstuvwxyz1234567890"
	maxLength := 5

	for i := 0; i < b.N; i++ {
		_, err := svc.BruteForceMD5(hash, strings.Split(alphabet, ""), maxLength, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}
