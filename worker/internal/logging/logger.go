package logging

import (
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timandy/routine"
	"io"
	"math"
	"os"
	"time"
)

func Setup(env config.Env) {
	var writer io.Writer

	switch env {
	case config.EnvProd:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		zerolog.CallerSkipFrameCount = math.MaxInt32
		zerolog.TimestampFieldName = "@timestamp"
		writer = os.Stdout

	case config.EnvDev:
		fallthrough

	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		writer = zerolog.NewConsoleWriter(
			func(w *zerolog.ConsoleWriter) {
				w.TimeFormat = time.RFC3339
			},
		)
	}

	log.Logger = zerolog.
		New(writer).
		Hook(gidHook{}).
		With().
		Timestamp().
		Caller().
		Logger()
}

type gidHook struct{}

func (h gidHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	e.Int64("gid", routine.Goid())
}
