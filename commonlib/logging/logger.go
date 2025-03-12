package logging

import (
	"io"
	"math"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timandy/routine"
)

func Setup(isDev bool) {
	var writer io.Writer

	if isDev {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		writer = zerolog.NewConsoleWriter(
			func(w *zerolog.ConsoleWriter) {
				w.TimeFormat = time.RFC3339
			},
		)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		zerolog.CallerSkipFrameCount = math.MaxInt32
		zerolog.TimestampFieldName = "@timestamp"
		writer = os.Stdout
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
