package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type HashCrackTask struct {
	ObjectID   primitive.ObjectID  `bson:"_id"`
	Hash       string              `bson:"hash"`
	MaxLength  int                 `bson:"maxLength"`
	PartCount  int                 `bson:"partCount"`
	Status     HashCrackTaskStatus `bson:"status"`
	Reason     *string             `bson:"reason,omitempty"`
	FinishedAt *time.Time          `bson:"finishedAt,omitempty"`
	CreatedAt  time.Time           `bson:"createdAt"`
	UpdatedAt  time.Time           `bson:"updatedAt"`
}

type HashCrackTaskWithSubtasks struct {
	ObjectID   primitive.ObjectID  `bson:"_id"`
	Hash       string              `bson:"hash"`
	MaxLength  int                 `bson:"maxLength"`
	PartCount  int                 `bson:"partCount"`
	Status     HashCrackTaskStatus `bson:"status"`
	Reason     *string             `bson:"reason,omitempty"`
	FinishedAt *time.Time          `bson:"finishedAt,omitempty"`
	CreatedAt  time.Time           `bson:"createdAt"`
	UpdatedAt  time.Time           `bson:"updatedAt"`
	Subtasks   []*HashCrackSubtask `bson:"subtasks,omitempty"`
}

func (c *HashCrackTaskWithSubtasks) ToHashCrackTask() *HashCrackTask {
	return &HashCrackTask{
		ObjectID:   c.ObjectID,
		Hash:       c.Hash,
		MaxLength:  c.MaxLength,
		PartCount:  c.PartCount,
		Status:     c.Status,
		Reason:     c.Reason,
		FinishedAt: c.FinishedAt,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

type HashCrackTaskStatus string

const (
	HashCrackTaskStatusPending      HashCrackTaskStatus = "PENDING"
	HashCrackTaskStatusInProgress   HashCrackTaskStatus = "IN_PROGRESS"
	HashCrackTaskStatusPartialReady HashCrackTaskStatus = "PARTIAL_READY"
	HashCrackTaskStatusReady        HashCrackTaskStatus = "READY"
	HashCrackTaskStatusError        HashCrackTaskStatus = "ERROR"
	HashCrackTaskStatusUnknown      HashCrackTaskStatus = "UNKNOWN"
)

func (c HashCrackTaskStatus) String() string {
	return string(c)
}

func ParseHashCrackTaskStatus(s string) HashCrackTaskStatus {
	switch s {
	case "PENDING":
		return HashCrackTaskStatusPending
	case "IN_PROGRESS":
		return HashCrackTaskStatusInProgress
	case "PARTIAL_READY":
		return HashCrackTaskStatusPartialReady
	case "READY":
		return HashCrackTaskStatusReady
	case "ERROR":
		return HashCrackTaskStatusError
	default:
		return HashCrackTaskStatusUnknown
	}
}
