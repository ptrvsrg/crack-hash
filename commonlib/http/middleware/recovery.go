package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/ptrvsrg/crack-hash/commonlib/http/types"
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
			errOutput := types.ErrorOutput{
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
