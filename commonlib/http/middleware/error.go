package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	"github.com/ptrvsrg/crack-hash/commonlib/http/types"
)

func ErrorMiddleware() gin.HandlerFunc {
	log.Debug().Msg("setup error middleware")
	logger := log.With().Str("middleware", "error").Logger()

	return func(c *gin.Context) {
		c.Next()

		// get the last error
		lastErr := c.Errors.Last()
		if lastErr == nil {
			return
		}

		logger.Debug().Msg("found the last error")
		err := lastErr.Err
		logger.Error().Err(err).Stack().Msg("failed to handle request")

		// build error response
		logger.Debug().Msg("build error response")
		status := c.Writer.Status()
		errOutput := types.ErrorOutput{
			Timestamp: time.Now(),
			Message:   err.Error(),
			Status:    status,
			Path:      c.Request.URL.Path,
		}

		var (
			validErrs validator.ValidationErrors
			syntaxErr *json.SyntaxError
		)
		switch {
		case errors.As(err, &validErrs):
			errOutput.Message = errors.Join(validErrs).Error()

		case errors.As(err, &syntaxErr):
			errOutput.Message = "invalid json"

		case errors.Is(err, io.EOF):
			errOutput.Message = "empty body"

		case status == http.StatusInternalServerError:
			errOutput.Message = "internal server error"
		}

		// send error response
		contentType := c.Request.Header.Get(headers.ContentType)
		if contentType == gin.MIMEXML {
			c.XML(status, errOutput)
		} else {
			c.JSON(status, errOutput)
		}
	}
}
