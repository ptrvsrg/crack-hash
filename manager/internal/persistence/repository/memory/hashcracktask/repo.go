package hashcracktask

import (
	"context"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"

	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
)

type repo struct {
	crackTasks map[string]*entity.HashCrackTask
	rw         sync.RWMutex
	logger     zerolog.Logger
}

func NewRepo() repository.HashCrackTask {
	return &repo{
		crackTasks: make(map[string]*entity.HashCrackTask),
		logger:     log.With().Str("repo", "hash-crack").Str("type", "memory").Logger(),
	}
}

func (r *repo) GetAllByHashAndMaxLength(_ context.Context, hash string, maxLength int) (
	[]*entity.HashCrackTask, error,
) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	r.logger.Debug().
		Str("hash", hash).
		Int("max-length", maxLength).
		Msg("get all pending crack tasks")

	values := maps.Values(r.crackTasks)
	sort.Slice(
		values, func(i, j int) bool {
			return values[i].CreatedAt.Before(values[j].CreatedAt)
		},
	)

	return lo.Filter(
		values,
		func(task *entity.HashCrackTask, _ int) bool {
			return task.Status == entity.HashCrackTaskStatusInProgress &&
				task.Hash == hash &&
				task.MaxLength == maxLength
		},
	), nil
}

func (r *repo) CountByStatus(_ context.Context, status entity.HashCrackTaskStatus) (int, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	r.logger.Debug().
		Str("status", status.String()).
		Msg("count crack tasks by status")

	values := maps.Values(r.crackTasks)
	sort.Slice(
		values, func(i, j int) bool {
			return values[i].CreatedAt.Before(values[j].CreatedAt)
		},
	)

	return lo.CountBy(
		values,
		func(task *entity.HashCrackTask) bool {
			return task.Status == status
		},
	), nil
}

func (r *repo) GetAllFinished(_ context.Context) ([]*entity.HashCrackTask, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	r.logger.Debug().Msg("get all finished crack tasks")

	values := maps.Values(r.crackTasks)
	sort.Slice(
		values, func(i, j int) bool {
			return values[i].CreatedAt.Before(values[j].CreatedAt)
		},
	)

	return lo.Filter(
		values,
		func(task *entity.HashCrackTask, _ int) bool {
			return task.Status == entity.HashCrackTaskStatusInProgress &&
				task.FinishedAt != nil &&
				task.FinishedAt.Before(time.Now())
		},
	), nil
}

func (r *repo) Get(_ context.Context, id string) (*entity.HashCrackTask, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	r.logger.Debug().Str("id", id).Msg("get crack task")

	task, ok := r.crackTasks[id]
	if !ok {
		return nil, repository.ErrCrackTaskNotFound
	}

	return task, nil
}

func (r *repo) Create(_ context.Context, task *entity.HashCrackTask) error {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.logger.Debug().Str("id", task.ID).Msg("insert crack task")

	if reflect.ValueOf(task).IsNil() {
		return repository.ErrCrackTaskIsNil
	}

	if _, ok := r.crackTasks[task.ID]; ok {
		return repository.ErrCrackTaskExists
	}

	r.crackTasks[task.ID] = task

	return nil
}

func (r *repo) Update(_ context.Context, task *entity.HashCrackTask) error {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.logger.Debug().Str("id", task.ID).Msg("update crack task")

	if reflect.ValueOf(task).IsNil() {
		return repository.ErrCrackTaskIsNil
	}

	if _, ok := r.crackTasks[task.ID]; !ok {
		return repository.ErrCrackTaskNotFound
	}

	r.crackTasks[task.ID] = task

	return nil
}

func (r *repo) DeleteAllExpired(_ context.Context, maxAge time.Duration) error {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.logger.Debug().
		Dur("max-age", maxAge).
		Msg("delete all expired crack tasks")

	for id, task := range r.crackTasks {
		if time.Since(task.CreatedAt) > maxAge {
			delete(r.crackTasks, id)
		}
	}

	return nil
}
