package domain

import (
	"context"

	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
)

type HashCrackTask interface {
	ExecuteTask(ctx context.Context, input *message.HashCrackTaskStarted) error
}

type Services struct {
	HashCrackTask HashCrackTask
}
