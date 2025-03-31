package hashcrack

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/maps"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

type svc struct {
	logger    zerolog.Logger
	cfg       config.TaskConfig
	taskRepo  repository.HashCrackTask
	splitSvc  infrastructure.TaskSplit
	publisher publisher.Publisher[message.HashCrackTaskStarted]
}

func NewService(
	logger zerolog.Logger,
	cfg config.TaskConfig,
	taskRepo repository.HashCrackTask,
	splitSvc infrastructure.TaskSplit,
	publisher publisher.Publisher[message.HashCrackTaskStarted],
) domain.HashCrackTask {

	return &svc{
		logger: logger.With().
			Str("type", "domain").
			Str("service", "hash-crack").
			Logger(),
		cfg:       cfg,
		taskRepo:  taskRepo,
		splitSvc:  splitSvc,
		publisher: publisher,
	}
}

func (s *svc) CreateTask(ctx context.Context, input *model.HashCrackTaskInput) (*model.HashCrackTaskIDOutput, error) {
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

	// Split task
	partCount, err := s.splitSvc.Split(ctx, input.MaxLength, len(s.cfg.Alphabet))
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

func (s *svc) GetTaskStatus(ctx context.Context, id string) (*model.HashCrackTaskStatusOutput, error) {
	s.logger.Info().Str("id", id).Msg("get task status")

	// Validate ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to validate ID")
		return nil, domain.ErrInvalidRequestID
	}

	// Get task
	task, err := s.taskRepo.Get(ctx, objID)
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

func (s *svc) SaveResultSubtask(ctx context.Context, input *message.HashCrackTaskResult) error {
	s.logger.Info().
		Str("id", input.RequestID).
		Int("part_number", input.PartNumber).
		Msg("save result subtask")

	// Validate ID
	objID, err := primitive.ObjectIDFromHex(input.RequestID)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to validate ID")
		return domain.ErrInvalidRequestID
	}

	// Get task
	task, err := s.taskRepo.Get(ctx, objID)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get task")

		if errors.Is(err, repository.ErrCrackTaskNotFound) {
			return domain.ErrTaskNotFound
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Update task
	task.Subtasks[input.PartNumber] = buildSubtaskEntity(input)

	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task")
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Check if task is finished
	s.logger.Debug().Msg("check if task is finished")

	hasSuccess, hasError, hasInProgress := s.hasSubtaskStatuses(task)
	if !hasInProgress && len(task.Subtasks) == task.PartCount {
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
	s.logger.Debug().Str("id", task.ObjectID.Hex()).Msg("start execute task")

	// Send tasks to workers
	for i := 0; i < task.PartCount; i++ {
		msg := buildTaskMessage(task, i, s.cfg.Alphabet)
		if err := s.publisher.SendMessage(ctx, msg, publisher.Persistent, false, false); err != nil {
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
	s.logger.Debug().Str("id", task.ObjectID.Hex()).Msg("finish task")

	// Mark task as ERROR
	s.markTaskAsErrorWithReason(task, "timeout")

	// Update tasks
	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task")
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *svc) hasSubtaskStatuses(task *entity.HashCrackTask) (hasSuccess, hasError, hasInProgress bool) {
	for _, subtask := range task.Subtasks {
		switch subtask.Status {
		case entity.HashCrackSubtaskStatusSuccess:
			hasSuccess = true
		case entity.HashCrackSubtaskStatusError:
			hasError = true
		case entity.HashCrackSubtaskStatusInProgress:
			hasInProgress = true
		case entity.HashCrackSubtaskStatusUnknown:
		}
	}

	return
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

func buildTaskIDOutput(task *entity.HashCrackTask) *model.HashCrackTaskIDOutput {
	return &model.HashCrackTaskIDOutput{
		RequestID: task.ObjectID.Hex(),
	}
}

func buildTaskEntity(input *model.HashCrackTaskInput, partCount int) *entity.HashCrackTask {
	return &entity.HashCrackTask{
		ObjectID:   primitive.NewObjectID(),
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

func buildSubtaskEntity(input *message.HashCrackTaskResult) *entity.HashCrackSubtask {
	if input.Status == entity.HashCrackSubtaskStatusError.String() {
		return &entity.HashCrackSubtask{
			PartNumber: input.PartNumber,
			Status:     entity.HashCrackSubtaskStatusError,
			Data:       []string{},
			Reason:     input.Error,
		}
	}

	data := make([]string, 0)
	percent := 0.0

	if input.Answer != nil {
		data = input.Answer.Words
		percent = input.Answer.Percent
	}

	return &entity.HashCrackSubtask{
		PartNumber: input.PartNumber,
		Status:     entity.ParseHashCrackSubtaskStatus(input.Status),
		Data:       data,
		Percent:    percent,
	}
}

func buildTaskStatusOutput(task *entity.HashCrackTask) *model.HashCrackTaskStatusOutput {
	data := make([]string, 0)
	percent := 0.0

	for _, subtask := range task.Subtasks {
		if task.PartCount > 0 {
			percent += subtask.Percent / float64(task.PartCount)
		}

		if task.Status != entity.HashCrackTaskStatusError {
			data = append(data, subtask.Data...)
		}
	}

	return &model.HashCrackTaskStatusOutput{
		Status:  task.Status.String(),
		Data:    data,
		Percent: math.Min(100.0, percent),
	}
}

func buildTaskMessage(task *entity.HashCrackTask, i int, alphabet string) *message.HashCrackTaskStarted {
	symbols := strings.Split(alphabet, "")

	return &message.HashCrackTaskStarted{
		RequestID:  task.ObjectID.Hex(),
		Hash:       task.Hash,
		MaxLength:  task.MaxLength,
		Alphabet:   message.Alphabet{Symbols: symbols},
		PartNumber: i,
		PartCount:  task.PartCount,
	}
}
