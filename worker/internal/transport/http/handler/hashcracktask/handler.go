package hashcracktask

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ptrvsrg/crack-hash/commonlib/http/handler"
	"github.com/ptrvsrg/crack-hash/commonlib/http/helper"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/pkg/model"
)

type hdlr struct {
	logger zerolog.Logger
	svc    domain.HashCrackTask
}

func NewHandler(svc domain.HashCrackTask) handler.Handler {
	return &hdlr{
		logger: log.With().Str("handler", "hash-crack-task").Logger(),
		svc:    svc,
	}
}

func (h *hdlr) RegisterRoutes(r *gin.Engine) {
	h.logger.Debug().Msg("register routes")

	api := r.Group("/internal/api/worker/hash/crack/task")
	{
		api.POST("", h.handleHashCrackTask)
	}
}

// handleHashCrackTask godoc
//
//	@Id				hashCrackTask
//	@Summary		Hash crack task
//	@Description	Request for executing hash crack task.
//	@Tags			Hash Crack Task API
//	@Accept			application/xml
//	@Produce		application/xml
//	@Param			input	body	model.HashCrackTaskInput	true	"Hash crack task input"
//	@Success		202
//	@Failure		400	{object}	model.ErrorOutput
//	@Failure		500	{object}	model.ErrorOutput
//	@Router			/internal/api/worker/hash/crack/task [post]
func (h *hdlr) handleHashCrackTask(ctx *gin.Context) {
	h.logger.Debug().Msg("handle hash crack task")

	input := &model.HashCrackTaskInput{}
	if err := ctx.ShouldBindXML(input); err != nil {
		_ = helper.ErrorWithStatus(ctx, http.StatusBadRequest, err)
		return
	}

	err := h.svc.ExecuteTask(ctx, input)
	if err != nil {
		_ = helper.ErrorWithStatus(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(202)
}
