package hashcrack

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/ptrvsrg/crack-hash/manager/internal/helper"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/transport/http/handler"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

var (
	ErrRequestIDNotFound = errors.New("requestID not found")
)

type hdlr struct {
	logger zerolog.Logger
	svc    domain.HashCrackTask
}

func NewHandler(svc domain.HashCrackTask) handler.Handler {
	return &hdlr{
		logger: log.With().Str("handler", "hash-crack").Logger(),
		svc:    svc,
	}
}

func (h *hdlr) RegisterRoutes(r *gin.Engine) {
	h.logger.Debug().Msg("register routes")

	exAPI := r.Group("/api/manager/hash/crack")
	{
		exAPI.POST("", h.handleHashCrack)
		exAPI.GET("/status", h.handleCheckHashCrackStatus)
	}

	inAPI := r.Group("/internal/api/manager/hash/crack")
	{
		inAPI.POST("/webhook", h.handleHashCrackTaskWebhook)
	}
}

// handleHashCrack godoc
//
//	@Id				HashCrack
//	@Summary	    Create new hash crack task
//	@Description	Request for create new hash crack task
//	@Tags			Hash Crack API
//	@Accept			application/json
//	@Produce		application/json
//	@Param			input	body	model.HashCrackTaskInput	true	"Hash crack task input"
//	@Success		202 {object} model.HashCrackTaskIDOutput
//	@Failure		400 {object} model.ErrorOutput
//	@Failure		500 {object} model.ErrorOutput
//	@Router			/api/manager/hash/crack [post]
func (h *hdlr) handleHashCrack(ctx *gin.Context) {
	h.logger.Debug().Msg("handle crack hash")

	input := &model.HashCrackTaskInput{}
	if err := ctx.ShouldBindJSON(input); err != nil {
		_ = helper.ErrorWithStatus(ctx, http.StatusBadRequest, err)
		return
	}

	output, err := h.svc.CreateTask(ctx, input)
	if err != nil {
		_ = helper.ErrorWithStatus(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(202, output)
}

// handleCheckHashCrackStatus godoc
//
//	@Id				CheckHashCrackStatus
//	@Summary	    Get status of hash crack task
//	@Description	Request for getting status of hash crack task
//	@Tags			Hash Crack API
//	@Produce		application/json
//	@Param			requestID	query	string	true	"Hash crack task ID"
//	@Success		200 {object} model.HashCrackTaskStatusOutput
//	@Failure		400 {object} model.ErrorOutput
//	@Failure		404 {object} model.ErrorOutput
//	@Failure		500 {object} model.ErrorOutput
//	@Router			/api/manager/hash/crack/status [get]
func (h *hdlr) handleCheckHashCrackStatus(c *gin.Context) {
	h.logger.Debug().Msg("handle check hash crack status")

	id, ok := c.GetQuery("requestID")
	if !ok {
		_ = helper.ErrorWithStatus(c, http.StatusBadRequest, ErrRequestIDNotFound)
		return
	}

	output, err := h.svc.GetTaskStatus(c, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrCrackTaskNotFound):
			_ = helper.ErrorWithStatus(c, http.StatusNotFound, err)
		default:
			_ = helper.ErrorWithStatus(c, http.StatusInternalServerError, err)
		}

		return
	}

	c.JSON(200, output)
}

// handleHashCrackTaskWebhook godoc
//
//	@Id				HashCrackTaskWebhook
//	@Summary	    Get status of hash crack task
//	@Description	Request for getting status of hash crack task
//	@Tags			Hash Crack API
//	@Accept			application/xml
//	@Produce		application/xml
//	@Param			input	body	model.HashCrackTaskWebhookInput	true	"Hash crack task webhook input"
//	@Success		200
//	@Failure		400 {object} model.ErrorOutput
//	@Failure		500 {object} model.ErrorOutput
//	@Router			/internal/api/manager/hash/crack/webhook [post]
func (h *hdlr) handleHashCrackTaskWebhook(ctx *gin.Context) {
	h.logger.Debug().Msg("handle hash crack task webhook")

	input := &model.HashCrackTaskWebhookInput{}
	if err := ctx.ShouldBindXML(input); err != nil {
		_ = helper.ErrorWithStatus(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.SaveResultTask(ctx, input); err != nil {
		_ = helper.ErrorWithStatus(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(200)
}
