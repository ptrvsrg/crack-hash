package hashcrack

import (
	"fmt"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
	"strings"
	"time"
)

func hasSubtaskStatuses(task *entity.HashCrackTaskWithSubtasks) (bool, bool, bool, bool) {
	var (
		hasSuccess    = false
		hasError      = false
		hasInProgress = false
		hasPending    = false
	)
	for _, subtask := range task.Subtasks {
		switch subtask.Status {
		case entity.HashCrackSubtaskStatusSuccess:
			hasSuccess = true
		case entity.HashCrackSubtaskStatusError:
			hasError = true
		case entity.HashCrackSubtaskStatusInProgress:
			hasInProgress = true
		case entity.HashCrackSubtaskStatusPending:
			hasPending = true
		case entity.HashCrackSubtaskStatusUnknown:
		}
	}

	return hasSuccess, hasError, hasInProgress, hasPending
}

func markTaskAsError(task *entity.HashCrackTask, subtasks []*entity.HashCrackSubtask) {
	reason := lo.Reduce(
		subtasks, func(acc string, subtask *entity.HashCrackSubtask, _ int) string {
			if subtask.Reason == nil {
				return acc
			}
			return fmt.Sprintf("%s; %s", acc, *subtask.Reason)
		}, "",
	)

	if task.Reason != nil {
		reason = fmt.Sprintf("%s; %s", *task.Reason, reason)
	}

	task.Status = entity.HashCrackTaskStatusError
	task.Reason = lo.ToPtr(reason)
}

func markSubtaskAsError(task *entity.HashCrackSubtask) {
	task.Status = entity.HashCrackSubtaskStatusError
}

func markTaskAsErrorWithReason(task *entity.HashCrackTask, reason string) {
	if task.Reason != nil {
		reason = fmt.Sprintf("%s; %s", *task.Reason, reason)
	}

	task.Status = entity.HashCrackTaskStatusError
	task.Reason = lo.ToPtr(reason)
}

func markSubtaskAsErrorWithReason(task *entity.HashCrackSubtask, reason string) {
	if task.Reason != nil {
		reason = fmt.Sprintf("%s; %s", *task.Reason, reason)
	}

	task.Status = entity.HashCrackSubtaskStatusError
	task.Reason = lo.ToPtr(reason)
}

func markTaskAsInProgress(task *entity.HashCrackTask) {
	task.Status = entity.HashCrackTaskStatusInProgress
}

func markSubtaskAsInProgress(task *entity.HashCrackSubtask) {
	task.Status = entity.HashCrackSubtaskStatusInProgress
}

func markTaskAsPartialReady(task *entity.HashCrackTask) {
	task.Status = entity.HashCrackTaskStatusPartialReady
}

func markTaskAsReady(task *entity.HashCrackTask) {
	task.Status = entity.HashCrackTaskStatusReady
}

func buildTaskIDOutput(task *entity.HashCrackTask) *model.HashCrackTaskIDOutput {
	return &model.HashCrackTaskIDOutput{
		RequestID: task.ObjectID.Hex(),
	}
}

func buildTaskEntityWithSubtasks(input *model.HashCrackTaskInput, partCount int) *entity.HashCrackTaskWithSubtasks {
	task := &entity.HashCrackTaskWithSubtasks{
		ObjectID:   primitive.NewObjectID(),
		Hash:       input.Hash,
		MaxLength:  input.MaxLength,
		PartCount:  partCount,
		Status:     entity.HashCrackTaskStatusPending,
		Reason:     nil,
		FinishedAt: nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	task.Subtasks = buildSubtaskEntities(partCount, task.ObjectID)

	return task
}

func buildSubtaskEntities(partCount int, taskID primitive.ObjectID) []*entity.HashCrackSubtask {
	subtasks := make([]*entity.HashCrackSubtask, partCount)
	for i := 0; i < partCount; i++ {
		subtasks[i] = &entity.HashCrackSubtask{
			ObjectID:   primitive.NewObjectID(),
			TaskID:     taskID,
			PartNumber: i,
			Status:     entity.HashCrackSubtaskStatusPending,
			Data:       nil,
			Reason:     nil,
		}
	}

	return subtasks
}

func buildTaskStatusOutput(task *entity.HashCrackTaskWithSubtasks) *model.HashCrackTaskStatusOutput {
	allData := make([]string, 0)
	averagePercent := 0.0
	subtaskOutputs := make([]model.HashCrackSubtaskStatusOutput, len(task.Subtasks))

	for i, subtask := range task.Subtasks {
		if task.PartCount > 0 {
			averagePercent += subtask.Percent / float64(task.PartCount)
		}

		if task.Status != entity.HashCrackTaskStatusError {
			allData = append(allData, subtask.Data...)
		}

		subtaskOutputs[i] = model.HashCrackSubtaskStatusOutput{
			Status:  subtask.Status.String(),
			Data:    subtask.Data,
			Percent: subtask.Percent,
		}
	}

	return &model.HashCrackTaskStatusOutput{
		Status:   task.Status.String(),
		Data:     allData,
		Percent:  math.Min(100.0, averagePercent),
		Subtasks: subtaskOutputs,
	}
}

func buildTaskMessage(task *entity.HashCrackTask, i int, alphabet string) *message.HashCrackTaskStarted {
	symbols := strings.Split(alphabet, "")

	return &message.HashCrackTaskStarted{
		RequestID:  task.ObjectID.Hex(),
		Hash:       task.Hash,
		MaxLength:  task.MaxLength,
		Alphabet:   message.Alphabet{Symbols: symbols},
		PartNumber: i,
		PartCount:  task.PartCount,
	}
}

func partialUpdateSubtaskEntity(subtask *entity.HashCrackSubtask, input *message.HashCrackTaskResult) {
	subtask.Status = entity.ParseHashCrackSubtaskStatus(input.Status)
	subtask.Reason = input.Error

	if input.Answer != nil {
		subtask.Data = input.Answer.Words
		subtask.Percent = input.Answer.Percent
	}
}
