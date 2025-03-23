package middleware

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	requestIDContextKey = "request-id"
	requestIDHeaderKey  = "X-Request-ID"
	maxLogBodySize      = 1024
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("failed to write response: %w", err)
	}

	return n, nil
}

func LoggerMiddleware(ignorePathRegexpStrs ...string) gin.HandlerFunc {
	log.Debug().Msg("setup error middleware")
	logger := log.With().Str("middleware", "logger").Logger()

	ignorePathRegexps := make([]*regexp.Regexp, 0)
	for _, regexpStr := range ignorePathRegexpStrs {
		log.Debug().Msgf("compile ignore regexp: %s", regexpStr)
		compiledRegexp, err := regexp.Compile(regexpStr)
		if err != nil {
			log.Error().Err(err).Stack().Msgf("failed to compiled regexp")
			continue
		}

		ignorePathRegexps = append(ignorePathRegexps, compiledRegexp)
	}

	return func(c *gin.Context) {
		// Request
		start := time.Now()
		path := c.Request.URL.Path

		for _, pathRegexp := range ignorePathRegexps {
			match := pathRegexp.MatchString(path)
			if match {
				return
			}
		}

		query := c.Request.URL.RawQuery
		method := c.Request.Method
		host := c.Request.Host
		userAgent := c.Request.UserAgent()
		ip := c.ClientIP()

		params := map[string]string{}
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}

		requestID := c.GetHeader(requestIDHeaderKey)
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header(requestIDHeaderKey, requestID)
		}
		c.Set(requestIDContextKey, requestID)

		reqEvent := logger.Info().
			Str("id", requestID).
			Str("method", method).
			Str("host", host).
			Str("path", path).
			Str("query", query).
			Interface("params", params).
			Str("ip", ip).
			Str("user-agent", userAgent)
		if log.Logger.GetLevel() <= zerolog.DebugLevel {
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ := io.ReadAll(tee)
			c.Request.Body = io.NopCloser(&buf)

			reqEvent.Interface("headers", c.Request.Header)
			reqEvent.Str("body", string(body))

			c.Writer = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		}

		reqEvent.Msg("Incoming request")

		// Process
		c.Next()

		// Response
		latency := time.Since(start)
		status := c.Writer.Status()

		respEvent := logger.Info().
			Str("request-id", requestID).
			Str("method", method).
			Str("host", host).
			Str("path", path).
			Str("query", query).
			Interface("params", params).
			Str("ip", ip).
			Str("user-agent", userAgent).
			Dur("latency", latency).
			Int("status", status)
		if log.Logger.GetLevel() <= zerolog.DebugLevel {
			body := c.Writer.(*bodyLogWriter).body.String()
			if len(body) > maxLogBodySize {
				body = body[:maxLogBodySize] + "..."
			}

			respEvent.Interface("body", body)
		}

		respEvent.Msg("Outcoming response")
	}
}
