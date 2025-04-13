package taskwithsubtasks

import (
	"context"
	"fmt"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type svc struct {
	logger      zerolog.Logger
	taskRepo    repository.HashCrackTask
	subtaskRepo repository.HashCrackSubtask
}

func NewService(
	taskRepo repository.HashCrackTask,
	subtaskRepo repository.HashCrackSubtask,
) infrastructure.TaskWithSubtasks {
	return &svc{
		logger:      log.With().Str("type", "infrastructure").Str("service", "task-with-subtasks").Logger(),
		taskRepo:    taskRepo,
		subtaskRepo: subtaskRepo,
	}
}

func (s *svc) CreateTaskWithSubtasks(ctx context.Context, task *entity.HashCrackTaskWithSubtasks) error {
	s.logger.Debug().Str("id", task.ObjectID.Hex()).Msg("create task with subtasks")

	_, err := s.taskRepo.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		if err := s.taskRepo.Create(ctx, task.ToHashCrackTask()); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to create task")
			return nil, fmt.Errorf("failed to create task: %w", err)
		}

		if err := s.subtaskRepo.CreateAll(ctx, task.Subtasks); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to create subtasks")
			return nil, fmt.Errorf("failed to create subtasks: %w", err)
		}

		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("failed to create task and subtasks: %w", err)
	}

	return nil
}

func (s *svc) UpdateTaskWithSubtasks(ctx context.Context, task *entity.HashCrackTaskWithSubtasks) error {
	s.logger.Debug().Str("id", task.ObjectID.Hex()).Msg("update task with subtasks")

	_, err := s.taskRepo.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		if err := s.taskRepo.Update(ctx, task.ToHashCrackTask()); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to update task")
			return nil, fmt.Errorf("failed to update task: %w", err)
		}

		if err := s.subtaskRepo.UpdateAll(ctx, task.Subtasks); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to update subtasks")
			return nil, fmt.Errorf("failed to update subtasks: %w", err)
		}

		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("failed to update task and subtasks: %w", err)
	}

	return nil
}

func (s *svc) DeleteTasksWithSubtasks(ctx context.Context, tasks []*entity.HashCrackTaskWithSubtasks) error {
	s.logger.Debug().Int("count", len(tasks)).Msg("delete tasks with subtasks")

	// Aggregate task and subtask IDs
	taskIds := make([]primitive.ObjectID, len(tasks))
	subtaskIds := make([]primitive.ObjectID, 0)

	for i, task := range tasks {
		taskIds[i] = task.ObjectID

		localSubtaskIds := lo.Map(task.Subtasks, func(subtask *entity.HashCrackSubtask, _ int) primitive.ObjectID {
			return subtask.ObjectID
		})
		subtaskIds = append(subtaskIds, localSubtaskIds...)
	}

	_, err := s.taskRepo.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		// Delete tasks
		if err := s.taskRepo.DeleteAllByIDs(ctx, taskIds); err != nil {
			return nil, fmt.Errorf("failed to delete tasks: %w", err)
		}

		// Delete subtasks
		if err := s.subtaskRepo.DeleteAllByIDs(ctx, subtaskIds); err != nil {
			return nil, fmt.Errorf("failed to delete subtasks: %w", err)
		}

		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete tasks and subtasks: %w", err)
	}

	return nil
}
