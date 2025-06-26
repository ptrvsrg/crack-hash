package domain

import (
	"context"
	"errors"

	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

var (
	ErrTooManyTasks          = errors.New("too many tasks")
	ErrTaskNotFound          = errors.New("task not found")
	ErrSubtaskNotFound       = errors.New("subtask not found")
	ErrInvalidRequestID      = errors.New("invalid request ID")
	ErrTaskFinishedByTimeout = errors.New("task finished by timeout")
)

type HashCrackTask interface {
	CreateTask(ctx context.Context, input *model.HashCrackTaskInput) (*model.HashCrackTaskIDOutput, error)
	GetTaskMetadatas(ctx context.Context, limit, offset int) (*model.HashCrackTaskMetadatasOutput, error)
	GetTaskStatus(ctx context.Context, id string) (*model.HashCrackTaskStatusOutput, error)
	SaveResultSubtask(ctx context.Context, input *message.HashCrackTaskResult) error
	ExecutePendingSubtasks(ctx context.Context) error
	FinishTimeoutTasks(ctx context.Context) error
	DeleteExpiredTasks(ctx context.Context) error
}

type Health interface {
	Health(ctx context.Context) error
}

type Services struct {
	HashCrackTask HashCrackTask
	Health        Health
}
