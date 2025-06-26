package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func CorsMiddleware(cfg cors.Config) gin.HandlerFunc {
	log.Debug().Msg("setup cors middleware")

	return cors.New(cfg)
}
