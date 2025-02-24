package entity

import (
	"time"
)

type HashCrackTask struct {
	ID         string
	Hash       string
	MaxLength  int
	PartCount  int
	Subtasks   []*HashCrackSubtask
	Status     string
	Reason     *string
	FinishedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type HashCrackSubtask struct {
	PartNumber int
	Data       []string
}
