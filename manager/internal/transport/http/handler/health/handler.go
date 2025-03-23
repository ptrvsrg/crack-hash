package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/ptrvsrg/crack-hash/commonlib/http/handler"
	"github.com/ptrvsrg/crack-hash/commonlib/http/helper"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	_ "github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

type hdlr struct {
	logger zerolog.Logger
	svc    domain.Health
}

func NewHandler(logger zerolog.Logger, svc domain.Health) handler.Handler {
	return &hdlr{
		logger: logger.With().Str("handler", "health").Logger(),
		svc:    svc,
	}
}

func (h *hdlr) RegisterRoutes(r *gin.Engine) {
	h.logger.Debug().Msgf("register routes")

	api := r.Group("/api/manager/health")
	{
		api.GET("/readiness", h.handleHealthReadiness)
		api.GET("/liveness", h.handleHealthLiveness)
	}
}

// handleHealthReadiness godoc
//
//	@Id				healthReadiness
//	@Summary		Health readiness
//	@Description	Request for getting health readiness. In response will be status of all check (database, cache, message queue).
//	@Tags			Health API
//	@Produce		application/json
//	@Success		200
//	@Failure		503	{object}	model.ErrorOutput
//	@Router			/api/manager/health/readiness [get]
func (h *hdlr) handleHealthReadiness(ctx *gin.Context) {
	h.logger.Debug().Msg("handle health readiness")

	if err := h.svc.Health(ctx); err != nil {
		_ = helper.ErrorWithStatus(ctx, http.StatusServiceUnavailable, err)
		return
	}

	ctx.String(200, "OK")
}

// handleHealthLiveness godoc
//
//	@Id				healthLiveness
//	@Summary		Health liveness
//	@Description	Request for getting health liveness.
//	@Tags			Health API
//	@Success		200
//	@Router			/api/manager/health/liveness [get]
func (h *hdlr) handleHealthLiveness(ctx *gin.Context) {
	h.logger.Debug().Msg("handle health liveness")
	ctx.String(200, "OK")
}
