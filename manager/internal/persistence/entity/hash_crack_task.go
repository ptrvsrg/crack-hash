package entity

import (
	"time"
)

type HashCrackTask struct {
	ID         string
	Hash       string
	MaxLength  int
	PartCount  int
	Subtasks   map[int]*HashCrackSubtask
	Status     HashCrackTaskStatus
	Reason     *string
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
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
