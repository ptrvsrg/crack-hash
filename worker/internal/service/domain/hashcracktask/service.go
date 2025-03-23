package hashcracktask

import (
	"context"
	"errors"
	"fmt"

	jobqueue "github.com/dirkaholic/kyoo"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"resty.dev/v3"

	managermodel "github.com/ptrvsrg/crack-hash/manager/pkg/model"
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	workermodel "github.com/ptrvsrg/crack-hash/worker/pkg/model"
)

type job struct {
	handler func()
}

func newJob(handler func()) *job {
	return &job{
		handler: handler,
	}
}

func (j *job) Process() {
	j.handler()
}

type svc struct {
	logger     zerolog.Logger
	cfg        config.ManagerConfig
	client     *resty.Client
	jobQueue   *jobqueue.JobQueue
	bruteforce infrastructure.HashBruteForce
}

func NewService(
	cfg config.ManagerConfig,
	client *resty.Client,
	jobQueue *jobqueue.JobQueue,
	bruteforce infrastructure.HashBruteForce,
) domain.HashCrackTask {
	return &svc{
		logger: log.With().
			Str("type", "domain").
			Str("service", "hash-crack-task").
			Logger(),
		cfg:        cfg,
		client:     client,
		jobQueue:   jobQueue,
		bruteforce: bruteforce,
	}
}

func (s *svc) ExecuteTask(ctx context.Context, input *workermodel.HashCrackTaskInput) error {
	s.logger.Info().
		Str("id", input.RequestID).
		Int("part", input.PartNumber).
		Msg("start brute force md5")

	s.jobQueue.Submit(
		newJob(func() { s.executeTask(ctx, input) }),
	)

	return nil
}

func (s *svc) executeTask(ctx context.Context, input *workermodel.HashCrackTaskInput) {
	// Brute force
	s.logger.Info().
		Str("id", input.RequestID).
		Int("part", input.PartNumber).
		Msg("brute force md5")

	answers, err := s.bruteforce.BruteForceMD5(input.Hash, input.Alphabet.Symbols, input.MaxLength, input.PartNumber)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to brute force md5")
		err = fmt.Errorf("failed to brute force md5: %w", err) // nolint
	}

	// Send result
	s.logger.Debug().Msg("send result webhook")

	webhookInput := &managermodel.HashCrackTaskWebhookInput{}
	webhookErrOutput := &workermodel.ErrorOutput{}

	if err == nil {
		webhookInput = buildSuccessWebhookRequest(input.RequestID, input.PartNumber, answers)
	} else {
		webhookInput = buildErrorWebhookRequest(input.RequestID, input.PartNumber, err.Error())
	}

	url := fmt.Sprintf("%s/internal/api/manager/hash/crack/webhook", s.cfg.Address)
	resp, err := s.client.R().
		SetHeader(headers.ContentType, gin.MIMEXML).
		SetContext(ctx).
		SetBody(webhookInput).
		SetError(webhookErrOutput).
		Post(url)

	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to send result webhook")
		return
	}

	if resp.IsError() {
		err = errors.New(webhookErrOutput.Message) // nolint
		s.logger.Error().Err(err).Stack().Msg("failed to execute task")
	}

	s.logger.Info().
		Str("id", input.RequestID).
		Int("part", input.PartNumber).
		Msg("end brute force md5")
}

func buildErrorWebhookRequest(
	requestID string, partNumber int, error string,
) *managermodel.HashCrackTaskWebhookInput {

	return &managermodel.HashCrackTaskWebhookInput{
		RequestID:  requestID,
		PartNumber: partNumber,
		Error:      lo.ToPtr(error),
	}
}

func buildSuccessWebhookRequest(
	requestID string, partNumber int, answers []string,
) *managermodel.HashCrackTaskWebhookInput {

	return &managermodel.HashCrackTaskWebhookInput{
		RequestID:  requestID,
		PartNumber: partNumber,
		Answer: &managermodel.Answer{
			Words: answers,
		},
	}
}
