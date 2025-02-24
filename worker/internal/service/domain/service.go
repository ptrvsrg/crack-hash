package domain

import (
	"context"
	"github.com/ptrvsrg/crack-hash/worker/pkg/model"
)

type HashCrackTask interface {
	ExecuteTask(ctx context.Context, input *model.HashCrackTaskInput) error
}

type Services struct {
	HashCrackTask HashCrackTask
}
