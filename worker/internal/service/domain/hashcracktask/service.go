package hashcracktask

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
	managermodel "github.com/ptrvsrg/crack-hash/manager/pkg/model"
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	workermodel "github.com/ptrvsrg/crack-hash/worker/pkg/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"resty.dev/v3"
)

type svc struct {
	logger     zerolog.Logger
	cfg        config.ManagerConfig
	client     *resty.Client
	bruteforce infrastructure.HashBruteForce
}

func NewService(
	cfg config.ManagerConfig,
	client *resty.Client,
	bruteforce infrastructure.HashBruteForce,
) domain.HashCrackTask {
	return &svc{
		logger:     log.With().Str("service", "hash-crack-task").Logger(),
		cfg:        cfg,
		client:     client,
		bruteforce: bruteforce,
	}
}

func (s *svc) ExecuteTask(ctx context.Context, input *workermodel.HashCrackTaskInput) error {
	s.logger.Info().Msg("start brute force md5")

	go s.executeTask(ctx, input)

	return nil
}

func (s *svc) executeTask(ctx context.Context, input *workermodel.HashCrackTaskInput) {
	// Brute force
	s.logger.Debug().Msg("brute force md5")

	answers, err := s.bruteforce.BruteForceMD5(input.Hash, input.Alphabet.Symbols, input.MaxLength, input.PartNumber)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to brute force md5")
		err = fmt.Errorf("failed to brute force md5: %w", err) // nolint
	}

	// Send result
	s.logger.Debug().Msg("send result webhook")

	webhookInput := &managermodel.HashCrackTaskWebhookInput{
		RequestID:  input.RequestID,
		PartNumber: input.PartNumber,
		Answer: struct {
			Words []string `xml:"words"`
		}{
			Words: answers,
		},
	}
	webhookErrOutput := &workermodel.ErrorOutput{}

	url := fmt.Sprintf("http://%s/internal/api/manager/hash/crack/webhook", s.cfg.Address)
	resp, err := s.client.R().
		SetHeader(headers.ContentType, gin.MIMEXML).
		SetContext(ctx).
		SetBody(webhookInput).
		SetError(webhookErrOutput).
		Post(url)

	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to send result webhook")
	}

	if resp.IsError() {
		err = errors.New(webhookErrOutput.Message) // nolint
		s.logger.Error().Err(err).Stack().Msg("failed to execute task")
	}
}
