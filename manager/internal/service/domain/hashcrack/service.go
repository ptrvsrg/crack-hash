package hashcrack

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/converter/hashcrack"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	managermodel "github.com/ptrvsrg/crack-hash/manager/pkg/model"
	workermodel "github.com/ptrvsrg/crack-hash/worker/pkg/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
	"time"
)

var (
	alphabet = "abcdefghijklmnopqrstuvwxyz1234567890"
	timeout  = 1 * time.Hour
)

type svc struct {
	logger   zerolog.Logger
	cfg      config.WorkerConfig
	client   *resty.Client
	taskRepo repository.HashCrackTask
	splitSvc infrastructure.TaskSplit
}

func NewService(
	cfg config.WorkerConfig,
	client *resty.Client,
	taskRepo repository.HashCrackTask,
	splitSvc infrastructure.TaskSplit,
) domain.HashCrackTask {

	return &svc{
		logger:   log.With().Str("service", "hash-crack").Logger(),
		cfg:      cfg,
		client:   client,
		taskRepo: taskRepo,
		splitSvc: splitSvc,
	}
}

func (s *svc) CreateTask(ctx context.Context, input *managermodel.HashCrackTaskInput) (*managermodel.HashCrackTaskIDOutput, error) {
	s.logger.Info().Msg("create task")

	// Split task
	partCount, err := s.splitSvc.Split(ctx, input.MaxLength, len(alphabet))
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to split task")
		return nil, fmt.Errorf("failed to split task: %w", err)
	}

	// Create and save task
	task := hashcrack.ConvertManagerTaskInputToTaskEntity(input, timeout)
	task.PartCount = partCount

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to create task")
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Execute task
	s.logger.Debug().Msg("async execute task")

	go func() {
		if err := s.startExecuteTask(ctx, task); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to execute task")
		}
	}()

	return &managermodel.HashCrackTaskIDOutput{
		RequestID: task.ID,
	}, nil
}

func (s *svc) GetTaskStatus(ctx context.Context, id string) (*managermodel.HashCrackTaskStatusOutput, error) {
	s.logger.Info().Msg("get task status")

	// Get task
	task, err := s.taskRepo.Get(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get task")
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Convert task
	return hashcrack.ConvertTaskEntityToManagerTaskStatusOutput(task), nil
}

func (s *svc) StartExecuteTask(ctx context.Context, id string) (err error) {
	s.logger.Info().Msgf("execute task with ID: %s", id)

	// Get task
	s.logger.Debug().Msg("get task")

	task, err := s.taskRepo.Get(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get task")
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Execute task
	s.logger.Debug().Msg("execute task")

	return s.startExecuteTask(ctx, task)
}

func (s *svc) SaveResultTask(ctx context.Context, input *managermodel.HashCrackTaskWebhookInput) error {
	s.logger.Info().Msg("save result task")

	// Get task
	s.logger.Debug().Msg("get task")

	task, err := s.taskRepo.Get(ctx, input.RequestID)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get task")
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Add result to task and save
	s.logger.Debug().Msg("add result to task")

	task.Subtasks = append(
		task.Subtasks,
		hashcrack.ConvertManagerTaskWebhookInputToSubtaskEntity(input),
	)

	// Update task
	s.logger.Debug().Msg("update task")

	if err := s.taskRepo.Update(ctx, input.RequestID, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task")
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Check if task is finished
	s.logger.Debug().Msg("check if task is finished")

	if len(task.Subtasks) == task.PartCount {
		markTaskAsReady(task)

		// Update task
		s.logger.Debug().Msg("update task")

		if err := s.taskRepo.Update(ctx, input.RequestID, task); err != nil {
			s.logger.Error().Err(err).Stack().Msg("failed to update task")
			return fmt.Errorf("failed to update task: %w", err)
		}
	}

	return nil
}

func (s *svc) FinishTasks(ctx context.Context) error {
	s.logger.Info().Msg("finish timeout tasks")

	// Get tasks
	s.logger.Debug().Msg("get tasks")

	tasks, err := s.taskRepo.GetAllFinished(ctx)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get tasks")
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		s.logger.Debug().Msg("no finished tasks found")
		return nil
	}

	// Update tasks
	s.logger.Debug().Msg("update tasks")

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

func (s *svc) FinishTask(ctx context.Context, id string) error {
	s.logger.Info().Msg("finish timeout task")

	// Get task
	s.logger.Debug().Msg("get task")

	task, err := s.taskRepo.Get(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to get task")
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Update task
	s.logger.Debug().Msg("update task")

	return s.finishTask(ctx, task)
}

func (s *svc) startExecuteTask(ctx context.Context, task *entity.HashCrackTask) (err error) {
	// Defer mark task as ERROR
	defer func() {
		if err != nil {
			markTaskAsError(task, err.Error())
			if err := s.taskRepo.Update(ctx, task.ID, task); err != nil {
				s.logger.Error().Err(err).Stack().Msg("failed to mark task as ERROR")
			}
		}
	}()

	// Send tasks to workers
	s.logger.Debug().Msg("send task to workers")

	group, _ := errgroup.WithContext(ctx)
	for i := 0; i < task.PartCount; i++ {
		group.Go(func() error {
			workerInput := hashcrack.ConvertTaskEntityToWorkerTaskInput(task, i, alphabet)
			if err := s.sendTaskToWorker(ctx, s.cfg.Address, workerInput); err != nil {
				return fmt.Errorf("failed to send task to worker: %w", err)
			}

			return nil
		})
	}

	// Wait for all tasks to send
	if err := group.Wait(); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to send task to workers")
		return fmt.Errorf("failed to send task to workers: %w", err)
	}

	return nil
}

func (s *svc) finishTask(ctx context.Context, task *entity.HashCrackTask) error {
	// Mark task as ERROR
	s.logger.Debug().Msg("mark task as ERROR")

	markTaskAsError(task, "timeout")

	// Update tasks
	s.logger.Debug().Msg("update task")

	if err := s.taskRepo.Update(ctx, task.ID, task); err != nil {
		s.logger.Error().Err(err).Stack().Msg("failed to update task")
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *svc) sendTaskToWorker(ctx context.Context, address string, input *workermodel.HashCrackTaskInput) error {
	errOutput := &workermodel.ErrorOutput{}

	url := fmt.Sprintf("http://%s/internal/api/worker/hash/crack/task", address)
	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader(headers.ContentType, gin.MIMEXML).
		SetBody(input).
		SetError(errOutput).
		Post(url)

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

func markTaskAsError(entity *entity.HashCrackTask, reason string) {
	if entity.Reason != nil {
		reason = fmt.Sprintf("%s; %s", *entity.Reason, reason)
	}

	entity.Status = managermodel.HashCrackStatusError.String()
	entity.Reason = lo.ToPtr(reason)
}

func markTaskAsReady(entity *entity.HashCrackTask) {
	entity.Status = managermodel.HashCrackStatusReady.String()
}
