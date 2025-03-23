package repository

import (
	"context"
	"errors"
	"time"

	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
)

var (
	ErrCrackTaskIsNil    = errors.New("crack task is nil")
	ErrCrackTaskNotFound = errors.New("crack task not found")
	ErrCrackTaskExists   = errors.New("crack task already exists")
)

type HashCrackTask interface {
	GetAllByHashAndMaxLength(ctx context.Context, hash string, maxLength int) ([]*entity.HashCrackTask, error)
	CountByStatus(ctx context.Context, status entity.HashCrackTaskStatus) (int, error)
	GetAllFinished(ctx context.Context) ([]*entity.HashCrackTask, error)
	Get(ctx context.Context, id string) (*entity.HashCrackTask, error)
	Create(ctx context.Context, task *entity.HashCrackTask) error
	Update(ctx context.Context, task *entity.HashCrackTask) error
	DeleteAllExpired(ctx context.Context, maxAge time.Duration) error
}

type Repositories struct {
	HashCrackTask HashCrackTask
}
