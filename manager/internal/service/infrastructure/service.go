package infrastructure

import (
	"context"
	"errors"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
)

var (
	ErrInvalidAlphabetLength = errors.New("invalid alphabet length")
	ErrInvalidWordMaxLength  = errors.New("invalid word max length")
)

type TaskSplit interface {
	Split(ctx context.Context, wordMaxLength, alphabetLength int) (int, error)
}

type TaskWithSubtasks interface {
	CreateTaskWithSubtasks(ctx context.Context, task *entity.HashCrackTaskWithSubtasks) error
	UpdateTaskWithSubtasks(ctx context.Context, task *entity.HashCrackTaskWithSubtasks) error
	DeleteTasksWithSubtasks(ctx context.Context, tasks []*entity.HashCrackTaskWithSubtasks) error
}

type Services struct {
	TaskSplit        TaskSplit
	TaskWithSubtasks TaskWithSubtasks
}
