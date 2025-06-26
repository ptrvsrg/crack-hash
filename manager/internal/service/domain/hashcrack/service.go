package hashcrack

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"

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
	logger              zerolog.Logger
	cfg                 config.TaskConfig
	taskRepo            repository.HashCrackTask
	subtaskRepo         repository.HashCrackSubtask
	splitSvc            infrastructure.TaskSplit
	taskWithSubtasksSvc infrastructure.TaskWithSubtasks
	publisher           publisher.Publisher[message.HashCrackTaskStarted]
}

func NewService(
	logger zerolog.Logger,
	cfg config.TaskConfig,
	taskRepo repository.HashCrackTask,
	subtaskRepo repository.HashCrackSubtask,
	splitSvc infrastructure.TaskSplit,
	taskWithSubtasksSvc infrastructure.TaskWithSubtasks,
	publisher publisher.Publisher[message.HashCrackTaskStarted],
) domain.HashCrackTask {

	return &svc{
		logger: logger.With().
			Str("type", "domain").
			Str("service", "hash-crack").
			Logger(),
		cfg:                 cfg,
		taskRepo:            taskRepo,
		subtaskRepo:         subtaskRepo,
		splitSvc:            splitSvc,
		taskWithSubtasksSvc: taskWithSubtasksSvc,
		publisher:           publisher,
	}
}

func (s *svc) CreateTask(ctx context.Context, input *model.HashCrackTaskInput) (*model.HashCrackTaskIDOutput, error) {
	s.logger.Info().
		Str("hash", input.Hash).
		Int("max_length", input.MaxLength).
		Msg("create task")

	// Get same tasks
	sameTask, err := s.taskRepo.GetByHashAndMaxLength(ctx, input.Hash, input.MaxLength, false)
	if err != nil && !errors.Is(err, repository.ErrCrackTaskNotFound) {
		s.logger.Warn().Err(err).Msg("failed to get same tasks")
	}

	if sameTask != nil {
		s.logger.Info().Msg("same task already exists")
		return buildTaskIDOutput(sameTask.ToHashCrackTask()), nil
	}

	// Split task
	partCount, err := s.splitSvc.Split(ctx, input.MaxLength, len(s.cfg.Alphabet))
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to split task")
		return nil, fmt.Errorf("failed to split task: %w", err)
	}

	// Create and save task with subtasks
	task := buildTaskEntityWithSubtasks(input, partCount)
	if err := s.taskWithSubtasksSvc.CreateTaskWithSubtasks(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task with subtasks: %w", err)
	}

	// Start execute tasks
	go func() {
		_ = s.startExecuteTask(ctx, task)
	}()

	return buildTaskIDOutput(task.ToHashCrackTask()), nil
}

func (s *svc) GetTaskMetadatas(
	ctx context.Context, limit,
	offset int,
) (*model.HashCrackTaskMetadatasOutput, error) {
	s.logger.Info().Int("limit", limit).Int("offset", offset).Msg("get task metadatas")

	// Get tasks and count
	var (
		tasks []*entity.HashCrackTaskWithSubtasks
		count int64
	)
	group, ctx := errgroup.WithContext(ctx)

	group.Go(
		func() error {
			var err error
			tasks, err = s.taskRepo.GetAll(ctx, limit, offset, false)
			if err != nil {
				return fmt.Errorf("failed to get tasks: %w", err)
			}
			return nil
		},
	)

	group.Go(
		func() error {
			var err error
			count, err = s.taskRepo.CountAll(ctx)
			if err != nil {
				return fmt.Errorf("failed to count tasks: %w", err)
			}
			return nil
		},
	)

	if err := group.Wait(); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get tasks and count")
		return nil, fmt.Errorf("failed to get tasks and count: %w", err)
	}

	// Convert task metadatas
	return buildTaskMetadataOutputs(count, tasks), nil
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
	task, err := s.taskRepo.Get(ctx, objID, true)
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

	// Update subtask and check if task is finished
	_, err = s.taskRepo.WithTransaction(
		ctx, func(ctx context.Context) (any, error) {
			// Get task
			taskWithSubtasks, err := s.taskRepo.Get(ctx, objID, true)
			if err != nil {
				s.logger.Error().Err(err).Stack().Msg("failed to get task")

				if errors.Is(err, repository.ErrCrackTaskNotFound) {
					return nil, domain.ErrTaskNotFound
				}

				return nil, fmt.Errorf("failed to get task: %w", err)
			}

			// Check if task is finished by timeout
			if taskWithSubtasks.Reason != nil && *taskWithSubtasks.Reason == domain.ErrTaskFinishedByTimeout.Error() {
				s.logger.Error().Err(domain.ErrTaskFinishedByTimeout).Msg("task finished by timeout")
				return nil, domain.ErrTaskFinishedByTimeout
			}

			// Get subtask
			var (
				subtaskIdx int
				ok         = false
			)
			for i, st := range taskWithSubtasks.Subtasks {
				if st.PartNumber == input.PartNumber {
					subtaskIdx = i
					ok = true
					break
				}
			}
			if !ok {
				s.logger.Error().Msg("subtask not found")
				return nil, domain.ErrSubtaskNotFound
			}

			// Update subtask
			partialUpdateSubtaskEntity(taskWithSubtasks.Subtasks[subtaskIdx], input)
			if err := s.subtaskRepo.Update(ctx, taskWithSubtasks.Subtasks[subtaskIdx]); err != nil {
				s.logger.Error().Err(err).Stack().Msg("failed to update task")
				return nil, fmt.Errorf("failed to update task: %w", err)
			}

			// Check if task is finished
			s.logger.Debug().Msg("check if task is finished")

			task := taskWithSubtasks.ToHashCrackTask()
			hasSuccess, hasError, hasInProgress, hasPending := hasSubtaskStatuses(taskWithSubtasks)
			if !hasInProgress && !hasPending {
				switch {
				case hasError && hasSuccess:
					s.logger.Info().Msg("mark task as PARTIAL_READY")
					markTaskAsPartialReady(task)
				case hasError:
					s.logger.Info().Msg("mark task as ERROR")
					markTaskAsError(task, taskWithSubtasks.Subtasks)
				case hasSuccess:
					s.logger.Info().Msg("mark task as READY")
					markTaskAsReady(task)
				}

				// Update task
				if err := s.taskRepo.Update(ctx, task); err != nil {
					s.logger.Error().Err(err).Stack().Msg("failed to update task")
					return nil, fmt.Errorf("failed to update task: %w", err)
				}

				s.logger.Info().Msg("task is finished")
			}

			return nil, nil
		},
	)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update subtask and check if task is finished")
		return fmt.Errorf("failed to update subtask and check if task is finished: %w", err)
	}

	return nil
}

func (s *svc) ExecutePendingSubtasks(ctx context.Context) error {
	s.logger.Info().Msg("execute pending subtasks")

	// Get pending subtasks
	subtasks, err := s.subtaskRepo.GetAllByStatus(ctx, entity.HashCrackSubtaskStatusPending)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get pending subtasks")
		return fmt.Errorf("failed to get pending subtasks: %w", err)
	}

	if len(subtasks) == 0 {
		s.logger.Debug().Msg("no pending subtasks found")
		return nil
	}

	s.logger.Debug().Int("count", len(subtasks)).Msg("pending subtasks found")

	// Calculate parent tasks
	subtasksMap := make(map[primitive.ObjectID][]*entity.HashCrackSubtask)
	for _, subtask := range subtasks {
		subtasksMapItem, ok := subtasksMap[subtask.TaskID]
		if !ok {
			subtasksMapItem = make([]*entity.HashCrackSubtask, 0)
		}

		subtasksMapItem = append(subtasksMapItem, subtask)
		subtasksMap[subtask.TaskID] = subtasksMapItem
	}

	// Execute subtasks
	errs := make([]error, 0)
	for taskID, subtasks := range subtasksMap {
		// Get task
		task, err := s.taskRepo.Get(ctx, taskID, false)
		if err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to get task")
			errs = append(errs, fmt.Errorf("failed to get task: %w", err))
			continue
		}

		// Execute subtasks
		if err := s.startExecuteSubtasks(ctx, task.ToHashCrackTask(), subtasks); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to execute subtasks")
			errs = append(errs, fmt.Errorf("failed to execute subtasks: %w", err))
			continue
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to execute pending subtasks: %w", multierr.Combine(errs...))
	}

	return nil
}

func (s *svc) FinishTimeoutTasks(ctx context.Context) error {
	s.logger.Info().Msg("finish timeout tasks")

	// Get tasks
	tasks, err := s.taskRepo.GetAllFinished(ctx, true)
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
		if err := s.finishTask(ctx, task); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to finish timeout tasks: %w", multierr.Combine(errs...))
	}

	return nil
}

func (s *svc) DeleteExpiredTasks(ctx context.Context) error {
	s.logger.Info().Msg("delete expired tasks")

	// Get tasks
	tasks, err := s.taskRepo.GetAllExpired(ctx, s.cfg.MaxAge, true)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get tasks")
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		s.logger.Debug().Msg("no expired tasks found")
		return nil
	}

	s.logger.Debug().Int("count", len(tasks)).Msg("expired tasks found")

	// Delete tasks with subtasks
	if err := s.taskWithSubtasksSvc.DeleteTasksWithSubtasks(ctx, tasks); err != nil {
		return fmt.Errorf("failed to delete tasks with subtasks: %w", err)
	}

	return nil
}

func (s *svc) startExecuteTask(ctx context.Context, taskWithSubtasks *entity.HashCrackTaskWithSubtasks) error {
	s.logger.Debug().Str("id", taskWithSubtasks.ObjectID.Hex()).Msg("start execute task")

	// Send tasks to workers
	for i := 0; i < taskWithSubtasks.PartCount; i++ {
		s.logger.Debug().
			Str("id", taskWithSubtasks.Subtasks[i].ObjectID.Hex()).
			Msg("send message to worker")

		msg := buildTaskMessage(taskWithSubtasks.ToHashCrackTask(), i, s.cfg.Alphabet)
		err := s.publisher.SendMessage(ctx, msg, publisher.Persistent, false, false)

		if err == nil {
			s.logger.Debug().Msg("mark subtask as IN_PROGRESS")
			markSubtaskAsInProgress(taskWithSubtasks.Subtasks[i])
		} else {
			s.logger.Error().Err(err).Stack().Msg("failed to send message")

			s.logger.Debug().Msg("mark subtask as ERROR")
			markSubtaskAsErrorWithReason(taskWithSubtasks.Subtasks[i], err.Error())
		}

		if err := s.subtaskRepo.Update(ctx, taskWithSubtasks.Subtasks[i]); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to update subtask")
			return fmt.Errorf("failed to update subtask: %w", err)
		}
	}

	// Mark task as IN_PROGRESS
	task := taskWithSubtasks.ToHashCrackTask()
	s.logger.Debug().Msg("mark task as IN_PROGRESS")
	markTaskAsInProgress(task)

	// Update task with subtasks

	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task")
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *svc) startExecuteSubtasks(
	ctx context.Context, task *entity.HashCrackTask, subtasks []*entity.HashCrackSubtask,
) error {
	subtaskIds := lo.Map(
		subtasks, func(subtask *entity.HashCrackSubtask, _ int) string {
			return subtask.ObjectID.Hex()
		},
	)

	s.logger.Debug().
		Str("id", task.ObjectID.Hex()).
		Strs("subtasks", subtaskIds).
		Msg("start execute subtasks")

	// Send tasks to workers
	for _, subtask := range subtasks {
		s.logger.Debug().
			Str("id", subtask.ObjectID.Hex()).
			Msg("send message to worker")

		msg := buildTaskMessage(task, subtask.PartNumber, s.cfg.Alphabet)
		err := s.publisher.SendMessage(ctx, msg, publisher.Persistent, false, false)

		if err == nil {
			s.logger.Debug().Msg("mark subtask as IN_PROGRESS")
			markSubtaskAsInProgress(subtask)
		} else {
			s.logger.Error().Err(err).Stack().Msg("failed to send message")

			s.logger.Debug().Msg("mark subtask as ERROR")
			markSubtaskAsErrorWithReason(subtask, err.Error())
		}

		if err := s.subtaskRepo.Update(ctx, subtask); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to update subtask")
			return fmt.Errorf("failed to update subtask: %w", err)
		}
	}

	// Mark task as IN_PROGRESS
	s.logger.Debug().Msg("mark task as IN_PROGRESS")
	markTaskAsInProgress(task)

	// Update task with subtasks

	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task")
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *svc) finishTask(ctx context.Context, task *entity.HashCrackTaskWithSubtasks) error {
	s.logger.Debug().Str("id", task.ObjectID.Hex()).Msg("finish task")

	// Mark task as ERROR
	s.logger.Debug().Msg("mark task as ERROR")
	markTaskAsErrorWithReason(task.ToHashCrackTask(), domain.ErrTaskFinishedByTimeout.Error())

	// Mark subtasks as ERROR
	for i := range task.Subtasks {
		s.logger.Debug().Str("id", task.Subtasks[i].ObjectID.Hex()).Msg("mark subtask as ERROR")
		markSubtaskAsErrorWithReason(task.Subtasks[i], domain.ErrTaskFinishedByTimeout.Error())
	}

	// Update task with subtasks
	if err := s.taskWithSubtasksSvc.UpdateTaskWithSubtasks(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task with subtasks")
		return fmt.Errorf("failed to update task with subtasks: %w", err)
	}

	return nil
}
