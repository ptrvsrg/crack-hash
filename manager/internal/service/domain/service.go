package domain

import (
	"context"
	"errors"

	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

var (
	ErrTooManyTasks = errors.New("too many tasks")
	ErrTaskNotFound = errors.New("task not found")
)

type HashCrackTask interface {
	CreateTask(ctx context.Context, input *model.HashCrackTaskInput) (*model.HashCrackTaskIDOutput, error)
	GetTaskStatus(ctx context.Context, id string) (*model.HashCrackTaskStatusOutput, error)
	SaveResultSubtask(ctx context.Context, input *model.HashCrackTaskWebhookInput) error
	FinishTimeoutTasks(ctx context.Context) error
	DeleteExpiredTasks(ctx context.Context) error
}

type Services struct {
	HashCrackTask HashCrackTask
}
