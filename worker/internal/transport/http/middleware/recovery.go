package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ptrvsrg/crack-hash/worker/pkg/model"
	"github.com/rs/zerolog/log"
	"net/http"
	"runtime/debug"
	"time"
)

func RecoveryMiddleware() gin.HandlerFunc {
	log.Debug().Msg("setup recovery middleware")
	logger := log.With().Str("middleware", "recovery").Logger()

	return func(ctx *gin.Context) {
		defer func() {
			errRaw := recover()

			// get error
			err, ok := errRaw.(error)
			if !ok || err == nil {
				return
			}

			logger.Error().Msgf("catch panic: %s\n%s", err, string(debug.Stack()))

			// build error output
			errOutput := model.ErrorOutput{
				Timestamp: time.Now(),
				Status:    http.StatusInternalServerError,
				Path:      ctx.Request.URL.Path,
				Message:   "internal server error",
			}

			ctx.JSON(http.StatusInternalServerError, errOutput)
		}()

		ctx.Next()
	}
}
