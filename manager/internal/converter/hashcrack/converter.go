package hashcrack

import (
	"github.com/google/uuid"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	managermodel "github.com/ptrvsrg/crack-hash/manager/pkg/model"
	workermodel "github.com/ptrvsrg/crack-hash/worker/pkg/model"
	"github.com/samber/lo"
	"strings"
	"time"
)

func ConvertManagerTaskInputToTaskEntity(input *managermodel.HashCrackTaskInput, timeout time.Duration) *entity.HashCrackTask {
	return &entity.HashCrackTask{
		ID:         uuid.New().String(),
		Hash:       input.Hash,
		MaxLength:  input.MaxLength,
		PartCount:  0,
		Subtasks:   make([]*entity.HashCrackSubtask, 0),
		Status:     managermodel.HashCrackStatusInProgress.String(),
		Reason:     nil,
		FinishedAt: time.Now().Add(timeout),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func ConvertManagerTaskWebhookInputToSubtaskEntity(input *managermodel.HashCrackTaskWebhookInput) *entity.HashCrackSubtask {
	return &entity.HashCrackSubtask{
		PartNumber: input.PartNumber,
		Data:       input.Answer.Words,
	}
}

func ConvertTaskEntityToManagerTaskStatusOutput(task *entity.HashCrackTask) *managermodel.HashCrackTaskStatusOutput {
	data := make([]string, 0)
	if task.Status == managermodel.HashCrackStatusReady.String() {
		data = lo.FlatMap(task.Subtasks, func(subtask *entity.HashCrackSubtask, _ int) []string {
			return subtask.Data
		})
	}

	return &managermodel.HashCrackTaskStatusOutput{
		Status: managermodel.ParseHashCrackTaskStatus(task.Status),
		Data:   data,
	}
}

func ConvertTaskEntityToWorkerTaskInput(task *entity.HashCrackTask, i int, alphabet string) *workermodel.HashCrackTaskInput {
	symbols := strings.Split(alphabet, "")
	return &workermodel.HashCrackTaskInput{
		RequestID: task.ID,
		Hash:      task.Hash,
		MaxLength: task.MaxLength,
		Alphabet: struct {
			Symbols []string `xml:"Symbols"`
		}{
			Symbols: symbols,
		},
		PartNumber: i,
		PartCount:  task.PartCount,
	}
}
