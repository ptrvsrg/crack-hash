package hashcrack

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/ptrvsrg/crack-hash/commonlib/http/handler"
	"github.com/ptrvsrg/crack-hash/commonlib/http/helper"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

var (
	ErrRequestIDNotFound = errors.New("requestID not found")
)

type hdlr struct {
	logger zerolog.Logger
	svc    domain.HashCrackTask
}

func NewHandler(logger zerolog.Logger, svc domain.HashCrackTask) handler.Handler {
	return &hdlr{
		logger: logger.With().Str("handler", "hash-crack").Logger(),
		svc:    svc,
	}
}

func (h *hdlr) RegisterRoutes(r *gin.Engine) {
	h.logger.Debug().Msg("register routes")

	exAPI := r.Group("/v1/hash/crack")
	{
		exAPI.POST("", h.handleCreateTask)
		exAPI.GET("/metadatas", h.handleGetTaskMetadatas)
		exAPI.GET("/status", h.handleGetTaskStatus)
	}
}

// handleCreateTask godoc
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
//	@Router			/v1/hash/crack [post]
func (h *hdlr) handleCreateTask(ctx *gin.Context) {
	h.logger.Debug().Msg("handle create task")

	input := &model.HashCrackTaskInput{}
	if err := ctx.ShouldBindJSON(input); err != nil {
		_ = helper.ErrorWithStatus(ctx, http.StatusBadRequest, err)
		return
	}

	output, err := h.svc.CreateTask(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTooManyTasks):
			_ = helper.ErrorWithStatus(ctx, http.StatusTooManyRequests, err)
		default:
			_ = helper.ErrorWithStatus(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(202, output)
}

// handleGetTaskMetadatas godoc
//
//	@Id				GetTaskMetadatas
//	@Summary	    Get metadatas of hash crack tasks
//	@Description	Request for getting metadatas of hash crack tasks
//	@Tags			Hash Crack API
//	@Produce		application/json
//	@Param			limit	query	int	false	"Limit"
//	@Param			offset	query	int	false	"Offset"
//	@Success		200 {object} model.HashCrackTaskMetadatasOutput
//	@Failure		400 {object} model.ErrorOutput
//	@Failure		500 {object} model.ErrorOutput
//	@Router			/v1/hash/crack/metadatas [get]
func (h *hdlr) handleGetTaskMetadatas(c *gin.Context) {
	h.logger.Debug().Msg("handle get task metadatas")

	input := &model.HashCrackTaskMetadataInput{}
	if err := c.ShouldBindQuery(input); err != nil {
		_ = helper.ErrorWithStatus(c, http.StatusBadRequest, err)
		return
	}

	output, err := h.svc.GetTaskMetadatas(c, input.Limit, input.Offset)
	if err != nil {
		_ = helper.ErrorWithStatus(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, output)
}

// handleGetTaskStatus godoc
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
//	@Router			/v1/hash/crack/status [get]
func (h *hdlr) handleGetTaskStatus(c *gin.Context) {
	h.logger.Debug().Msg("handle get task status")

	id, ok := c.GetQuery("requestID")
	if !ok {
		_ = helper.ErrorWithStatus(c, http.StatusBadRequest, ErrRequestIDNotFound)
		return
	}

	output, err := h.svc.GetTaskStatus(c, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRequestID):
			_ = helper.ErrorWithStatus(c, http.StatusBadRequest, err)
		case errors.Is(err, domain.ErrTaskNotFound):
			_ = helper.ErrorWithStatus(c, http.StatusNotFound, err)
		default:
			_ = helper.ErrorWithStatus(c, http.StatusInternalServerError, err)
		}

		return
	}

	c.JSON(200, output)
}
