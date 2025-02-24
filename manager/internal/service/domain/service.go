package domain

import (
	"context"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

type HashCrackTask interface {
	CreateTask(ctx context.Context, input *model.HashCrackTaskInput) (*model.HashCrackTaskIDOutput, error)
	GetTaskStatus(ctx context.Context, id string) (*model.HashCrackTaskStatusOutput, error)
	StartExecuteTask(ctx context.Context, id string) error
	SaveResultTask(ctx context.Context, input *model.HashCrackTaskWebhookInput) error
	FinishTasks(ctx context.Context) error
	FinishTask(ctx context.Context, id string) error
}

type Services struct {
	HashCrackTask HashCrackTask
}
