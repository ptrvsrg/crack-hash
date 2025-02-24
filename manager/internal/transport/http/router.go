package http

import (
	"errors"
	"github.com/ptrvsrg/crack-hash/manager/docs"
	"github.com/ptrvsrg/crack-hash/manager/internal/di"
	middleware2 "github.com/ptrvsrg/crack-hash/manager/internal/transport/http/middleware"
	"github.com/ptrvsrg/crack-hash/manager/internal/version"
	"github.com/rs/zerolog/log"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

var (
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrRouteNotFound    = errors.New("route not found")

	ignorePathRegexps = []string{
		"/api/manager/health.*",
		"/api/manager/swagger.*",
	}
)

// SetupRouter godoc
//
//	@title						Crack Hash Manager API
//	@version					0.0.0
//	@description				API for Crack Hash Manager
//	@host						localhost:8080
//	@query.collection.format	multi
//	@accept						json
//	@produce					json
//	@tag.name					Hash Crack API
//	@tag.description			API for cracking hashes and checking results
//	@tag.name					Health API
//	@tag.description			API for health checks
//	@tag.name					Swagger API
//	@tag.description			API for getting swagger specification
//	@contact.name				Petrov Sergey
//	@contact.email				s.petrov1@g.nsu.ru
//	@license.name				Apache 2.0
//	@license.url				https://www.apache.org/licenses/LICENSE-2.0.html
//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
func SetupRouter(c *di.Container) http.Handler {
	// Setup swagger docs
	docs.SwaggerInfo.Version = version.AppVersion

	// Create a new router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Setup middlewares
	log.Info().Msg("setup middlewares")

	r.Use(middleware2.LoggerMiddleware(ignorePathRegexps...))
	r.Use(middleware2.RecoveryMiddleware())
	r.Use(middleware2.ErrorMiddleware())

	// Setup routes
	log.Info().Msg("setup routes")

	r.HandleMethodNotAllowed = true
	r.NoMethod(handleNoMethod)
	r.NoRoute(handleNoRoute)

	for _, handler := range c.Handlers {
		handler.RegisterRoutes(r)
	}

	// Print registered routes
	routes := r.Routes()
	sort.Slice(
		routes, func(i, j int) bool {
			return routes[i].Path < routes[j].Path
		},
	)
	for _, routeInfo := range routes {
		log.Info().Msgf("registered route: %6s %s", routeInfo.Method, routeInfo.Path)
	}

	return r
}

func handleNoMethod(ctx *gin.Context) {
	log.Debug().Msg("handle method not allowed")

	ctx.Status(http.StatusMethodNotAllowed)
	_ = ctx.Error(ErrMethodNotAllowed)
}

func handleNoRoute(ctx *gin.Context) {
	log.Debug().Msg("handle route not found")

	ctx.Status(http.StatusNotFound)
	_ = ctx.Error(ErrRouteNotFound)
}
