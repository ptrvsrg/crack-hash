package memory

import (
	"context"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"reflect"
	"sync"
	"time"
)

type repo struct {
	crackTasks map[string]*entity.HashCrackTask
	rw         sync.RWMutex
	logger     zerolog.Logger
}

func NewRepo() repository.HashCrackTask {
	return &repo{
		crackTasks: make(map[string]*entity.HashCrackTask),
		logger:     log.With().Str("repo", "hash-crack").Str("repo-type", "memory").Logger(),
	}
}

func (r *repo) GetAllFinished(_ context.Context) ([]*entity.HashCrackTask, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	r.logger.Debug().Msg("get all finished crack tasks")

	values := maps.Values(r.crackTasks)

	return lo.Filter(
		values,
		func(task *entity.HashCrackTask, _ int) bool {
			return task.Status == model.HashCrackStatusInProgress.String() && task.FinishedAt.Before(time.Now())
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

func (r *repo) Update(_ context.Context, id string, task *entity.HashCrackTask) error {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.logger.Debug().Str("id", id).Msg("update crack task")

	if reflect.ValueOf(task).IsNil() {
		return repository.ErrCrackTaskIsNil
	}

	if _, ok := r.crackTasks[id]; !ok {
		return repository.ErrCrackTaskNotFound
	}

	r.crackTasks[id] = task

	return nil
}

func (r *repo) Delete(_ context.Context, id string) error {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.logger.Debug().Str("id", id).Msg("delete crack task")

	if _, ok := r.crackTasks[id]; !ok {
		return repository.ErrCrackTaskNotFound
	}

	delete(r.crackTasks, id)

	return nil
}
