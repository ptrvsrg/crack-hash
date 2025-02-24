package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type HashCrackTask struct {
	ObjectID   bson.ObjectID             `bson:"_id"`
	Hash       string                    `bson:"hash"`
	MaxLength  int                       `bson:"maxLength"`
	PartCount  int                       `bson:"partCount"`
	Subtasks   map[int]*HashCrackSubtask `bson:"subtasks"`
	Status     HashCrackTaskStatus       `bson:"status"`
	Reason     *string                   `bson:"reason,omitempty"`
	FinishedAt *time.Time                `bson:"finishedAt,omitempty"`
	CreatedAt  time.Time                 `bson:"createdAt"`
	UpdatedAt  time.Time                 `bson:"updatedAt"`
}

type HashCrackTaskStatus string

const (
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
