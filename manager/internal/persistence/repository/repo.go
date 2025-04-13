package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
)

var (
	ErrCrackTaskNotFound    = errors.New("crack task not found")
	ErrCrackTaskExists      = errors.New("crack task already exists")
	ErrCrackSubtaskNotFound = errors.New("crack subtask not found")
	ErrCrackSubtaskExists   = errors.New("crack subtask already exists")
)

type Transactor interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)
}

type HashCrackTask interface {
	Transactor

	GetAllFinished(ctx context.Context, withSubtasks bool) ([]*entity.HashCrackTaskWithSubtasks, error)
	GetAllExpired(
		ctx context.Context, maxAge time.Duration, withSubtasks bool,
	) ([]*entity.HashCrackTaskWithSubtasks, error)
	Get(ctx context.Context, id primitive.ObjectID, withSubtasks bool) (*entity.HashCrackTaskWithSubtasks, error)
	GetByHashAndMaxLength(
		ctx context.Context, hash string, maxLength int, withSubtasks bool,
	) (*entity.HashCrackTaskWithSubtasks, error)
	Create(ctx context.Context, task *entity.HashCrackTask) error
	Update(ctx context.Context, task *entity.HashCrackTask) error
	DeleteAllByIDs(ctx context.Context, ids []primitive.ObjectID) error
}

type HashCrackSubtask interface {
	Transactor

	GetByTaskIDAndPartNumber(
		ctx context.Context, taskID primitive.ObjectID, partNumber int,
	) (*entity.HashCrackSubtask, error)
	GetAllByTaskID(ctx context.Context, taskID primitive.ObjectID) ([]*entity.HashCrackSubtask, error)
	GetAllByTaskIDs(ctx context.Context, taskIDs []primitive.ObjectID) ([]*entity.HashCrackSubtask, error)
	GetAllByStatus(ctx context.Context, status entity.HashCrackSubtaskStatus) ([]*entity.HashCrackSubtask, error)
	Create(ctx context.Context, task *entity.HashCrackSubtask) error
	CreateAll(ctx context.Context, tasks []*entity.HashCrackSubtask) error
	Update(ctx context.Context, task *entity.HashCrackSubtask) error
	UpdateAll(ctx context.Context, tasks []*entity.HashCrackSubtask) error
	DeleteAllByIDs(ctx context.Context, ids []primitive.ObjectID) error
}

type Repositories struct {
	HashCrackTask    HashCrackTask
	HashCrackSubtask HashCrackSubtask
}
