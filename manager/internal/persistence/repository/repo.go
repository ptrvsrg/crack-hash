package repository

import (
	"context"
	"errors"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
)

var (
	ErrCrackTaskIsNil    = errors.New("crack task is nil")
	ErrCrackTaskNotFound = errors.New("crack task not found")
	ErrCrackTaskExists   = errors.New("crack task already exists")
)

type HashCrackTask interface {
	GetAllFinished(ctx context.Context) ([]*entity.HashCrackTask, error)
	Get(ctx context.Context, id string) (*entity.HashCrackTask, error)
	Create(ctx context.Context, task *entity.HashCrackTask) error
	Update(ctx context.Context, id string, task *entity.HashCrackTask) error
	Delete(ctx context.Context, id string) error
}

type Repositories struct {
	HashCrackTask HashCrackTask
}
