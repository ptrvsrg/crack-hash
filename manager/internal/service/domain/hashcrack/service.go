package hashcrack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"resty.dev/v3"

	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	managermodel "github.com/ptrvsrg/crack-hash/manager/pkg/model"
	workermodel "github.com/ptrvsrg/crack-hash/worker/pkg/model"
)

var (
	alphabet = "abcdefghijklmnopqrstuvwxyz1234567890"
)

type svc struct {
	logger   zerolog.Logger
	cfg      config.TaskConfig
	client   *resty.Client
	taskRepo repository.HashCrackTask
	splitSvc infrastructure.TaskSplit
}

func NewService(
	cfg config.TaskConfig,
	client *resty.Client,
	taskRepo repository.HashCrackTask,
	splitSvc infrastructure.TaskSplit,
) domain.HashCrackTask {

	return &svc{
		logger: log.With().
			Str("type", "domain").
			Str("service", "hash-crack").
			Logger(),
		cfg:      cfg,
		client:   client,
		taskRepo: taskRepo,
		splitSvc: splitSvc,
	}
}

func (s *svc) CreateTask(
	ctx context.Context, input *managermodel.HashCrackTaskInput,
) (*managermodel.HashCrackTaskIDOutput, error) {
	s.logger.Info().
		Str("hash", input.Hash).
		Int("max_length", input.MaxLength).
		Msg("create task")

	// Get same tasks
	sameTasks, err := s.taskRepo.GetAllByHashAndMaxLength(ctx, input.Hash, input.MaxLength)
	if err != nil {
		s.logger.Warn().Err(err).Msg("failed to get same tasks")
	}

	if len(sameTasks) > 0 {
		s.logger.Info().Msg("same task already exists")
		return buildTaskIDOutput(sameTasks[0]), nil
	}

	// Count in progress tasks
	inProgressCount, err := s.taskRepo.CountByStatus(ctx, entity.HashCrackTaskStatusInProgress)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to count in progress tasks")
		return nil, fmt.Errorf("failed to count in progress tasks: %w", err)
	}

	if inProgressCount >= s.cfg.Limit {
		s.logger.Error().Err(domain.ErrTooManyTasks).Msg("failed to create task")
		return nil, domain.ErrTooManyTasks
	}

	// Split task
	partCount, err := s.splitSvc.Split(ctx, input.MaxLength, len(alphabet))
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to split task")
		return nil, fmt.Errorf("failed to split task: %w", err)
	}

	// Create and save task
	task := buildTaskEntity(input, partCount)

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to create task")
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Start execute tasks
	go func() {
		_ = s.startExecuteTask(ctx, task)
	}()

	return buildTaskIDOutput(task), nil
}

func (s *svc) GetTaskStatus(ctx context.Context, id string) (*managermodel.HashCrackTaskStatusOutput, error) {
	s.logger.Info().Str("id", id).Msg("get task status")

	// Get task
	task, err := s.taskRepo.Get(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get task")

		if errors.Is(err, repository.ErrCrackTaskNotFound) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Convert task
	return buildTaskStatusOutput(task), nil
}

func (s *svc) SaveResultSubtask(ctx context.Context, input *managermodel.HashCrackTaskWebhookInput) error {
	s.logger.Info().
		Str("id", input.RequestID).
		Int("part_number", input.PartNumber).
		Msg("save result subtask")

	// Get task
	task, err := s.taskRepo.Get(ctx, input.RequestID)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get task")
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Add result to task and save
	if input.Error != nil {
		task.Subtasks[input.PartNumber] = buildErrorSubtaskEntity(input)
	} else {
		task.Subtasks[input.PartNumber] = buildSuccessSubtaskEntity(input)
	}

	// Update task
	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task")
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Check if task is finished
	s.logger.Debug().Msg("check if task is finished")

	if len(task.Subtasks) == task.PartCount {
		hasError := false
		hasSuccess := false

		for _, subtask := range task.Subtasks {
			if subtask.Status == entity.HashCrackSubtaskStatusSuccess {
				hasSuccess = true
			} else {
				hasError = true
			}
		}

		switch {
		case hasError && hasSuccess:
			s.markTaskAsPartialReady(task)
		case hasError:
			s.markTaskAsError(task)
		case hasSuccess:
			s.markTaskAsReady(task)
		}

		// Update task
		if err := s.taskRepo.Update(ctx, task); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to update task")
			return fmt.Errorf("failed to update task: %w", err)
		}

		s.logger.Info().Msg("task is finished")
	}

	return nil
}

func (s *svc) FinishTimeoutTasks(ctx context.Context) error {
	s.logger.Info().Msg("finish timeout tasks")

	// Get tasks
	tasks, err := s.taskRepo.GetAllFinished(ctx)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get tasks")
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		s.logger.Debug().Msg("no finished tasks found")
		return nil
	}

	s.logger.Debug().Int("count", len(tasks)).Msg("finished tasks found")

	// Finish tasks
	errs := make([]error, 0)
	for _, task := range tasks {
		err := s.finishTask(ctx, task)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to finish timeout tasks: %w", errors.Join(errs...))
	}

	return nil
}

func (s *svc) DeleteExpiredTasks(ctx context.Context) error {
	s.logger.Info().Msg("delete expired tasks")

	if err := s.taskRepo.DeleteAllExpired(ctx, s.cfg.MaxAge); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get tasks")
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	return nil
}

func (s *svc) startExecuteTask(ctx context.Context, task *entity.HashCrackTask) error {
	s.logger.Debug().Str("id", task.ID).Msg("start execute task")

	// Send tasks to workers
	for i := 0; i < task.PartCount; i++ {
		workerInput := buildWorkerRequest(task, i, alphabet)

		if err := s.sendTaskToWorker(ctx, workerInput); err != nil {
			task.Subtasks[i] = &entity.HashCrackSubtask{
				PartNumber: i,
				Status:     entity.HashCrackSubtaskStatusError,
				Data:       []string{},
				Reason:     lo.ToPtr(err.Error()),
			}

			if err := s.taskRepo.Update(ctx, task); err != nil {
				s.logger.Error().Err(err).Stack().Msg("failed to update task")
			}
		}
	}

	return nil
}

func (s *svc) finishTask(ctx context.Context, task *entity.HashCrackTask) error {
	s.logger.Debug().Str("id", task.ID).Msg("finish task")

	// Mark task as ERROR
	s.markTaskAsErrorWithReason(task, "timeout")

	// Update tasks
	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task")
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *svc) sendTaskToWorker(ctx context.Context, input *workermodel.HashCrackTaskInput) error {
	s.logger.Debug().Str("id", input.RequestID).Msg("send task to worker")

	errOutput := &workermodel.ErrorOutput{}

	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader(headers.ContentType, gin.MIMEXML).
		SetBody(input).
		SetError(errOutput).
		Post("/internal/api/worker/hash/crack/task")

	// Process response
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to send task to worker")
		return fmt.Errorf("failed to send task to worker: %w", err)
	}

	if resp.IsError() {
		err = errors.New(errOutput.Message) // nolint
		s.logger.Error().Err(err).Stack().Msg("failed to execute task")

		return fmt.Errorf("failed to execute task: %w", err)
	}

	return nil
}

func (s *svc) markTaskAsError(task *entity.HashCrackTask) {
	s.logger.Debug().Msg("mark task as ERROR")

	reason := lo.Reduce(
		maps.Values(task.Subtasks), func(acc string, subtask *entity.HashCrackSubtask, _ int) string {
			if subtask.Reason == nil {
				return acc
			}
			return fmt.Sprintf("%s; %s", acc, *subtask.Reason)
		}, "",
	)

	if task.Reason != nil {
		reason = fmt.Sprintf("%s; %s", *task.Reason, reason)
	}

	task.Status = entity.HashCrackTaskStatusError
	task.Reason = lo.ToPtr(reason)
}

func (s *svc) markTaskAsErrorWithReason(task *entity.HashCrackTask, reason string) {
	s.logger.Debug().Msg("mark task as ERROR")

	if task.Reason != nil {
		reason = fmt.Sprintf("%s; %s", *task.Reason, reason)
	}

	task.Status = entity.HashCrackTaskStatusError
	task.Reason = lo.ToPtr(reason)
}

func (s *svc) markTaskAsPartialReady(task *entity.HashCrackTask) {
	s.logger.Debug().Msg("mark task as PARTIAL READY")
	task.Status = entity.HashCrackTaskStatusPartialReady
}

func (s *svc) markTaskAsReady(task *entity.HashCrackTask) {
	s.logger.Debug().Msg("mark task as READY")
	task.Status = entity.HashCrackTaskStatusReady
}

func buildTaskIDOutput(task *entity.HashCrackTask) *managermodel.HashCrackTaskIDOutput {
	return &managermodel.HashCrackTaskIDOutput{
		RequestID: task.ID,
	}
}

func buildTaskEntity(input *managermodel.HashCrackTaskInput, partCount int) *entity.HashCrackTask {
	return &entity.HashCrackTask{
		ID:         uuid.New().String(),
		Hash:       input.Hash,
		MaxLength:  input.MaxLength,
		PartCount:  partCount,
		Subtasks:   make(map[int]*entity.HashCrackSubtask),
		Status:     entity.HashCrackTaskStatusInProgress,
		Reason:     nil,
		FinishedAt: nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func buildSuccessSubtaskEntity(input *managermodel.HashCrackTaskWebhookInput) *entity.HashCrackSubtask {
	data := make([]string, 0)
	if input.Answer != nil {
		data = input.Answer.Words
	}

	return &entity.HashCrackSubtask{
		PartNumber: input.PartNumber,
		Status:     entity.HashCrackSubtaskStatusSuccess,
		Data:       data,
	}
}

func buildErrorSubtaskEntity(input *managermodel.HashCrackTaskWebhookInput) *entity.HashCrackSubtask {
	return &entity.HashCrackSubtask{
		PartNumber: input.PartNumber,
		Status:     entity.HashCrackSubtaskStatusError,
		Data:       []string{},
		Reason:     input.Error,
	}
}

func buildTaskStatusOutput(task *entity.HashCrackTask) *managermodel.HashCrackTaskStatusOutput {
	data := make([]string, 0)
	if task.Status == entity.HashCrackTaskStatusReady || task.Status == entity.HashCrackTaskStatusPartialReady {
		values := maps.Values(task.Subtasks)

		data = lo.FlatMap(
			values, func(subtask *entity.HashCrackSubtask, _ int) []string {
				return subtask.Data
			},
		)
	}

	return &managermodel.HashCrackTaskStatusOutput{
		Status: task.Status.String(),
		Data:   data,
	}
}

func buildWorkerRequest(task *entity.HashCrackTask, i int, alphabet string) *workermodel.HashCrackTaskInput {
	symbols := strings.Split(alphabet, "")

	return &workermodel.HashCrackTaskInput{
		RequestID:  task.ID,
		Hash:       task.Hash,
		MaxLength:  task.MaxLength,
		Alphabet:   workermodel.Alphabet{Symbols: symbols},
		PartNumber: i,
		PartCount:  task.PartCount,
	}
}
