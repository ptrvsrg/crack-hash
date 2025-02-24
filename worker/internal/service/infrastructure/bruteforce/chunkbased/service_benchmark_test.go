package chunkbased_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/bruteforce/chunkbased"
)

func init() {
	logging.Setup(true)
}

func Benchmark(b *testing.B) {
	svc := chunkbased.NewService(10_000_000)
	hash := "abcde"
	alphabet := "abcdefghijklmnopqrstuvwxyz1234567890"
	maxLength := 5

	for i := 0; i < b.N; i++ {
		ch, err := svc.BruteForceMD5(hash, strings.Split(alphabet, ""), maxLength, 0, time.Second)
		if err != nil {
			b.Fatal(err)
		}

		for range ch {
		}
	}
}
