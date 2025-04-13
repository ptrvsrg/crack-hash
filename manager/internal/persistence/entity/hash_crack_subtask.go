package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type HashCrackSubtask struct {
	ObjectID   primitive.ObjectID     `bson:"_id"`
	TaskID     primitive.ObjectID     `bson:"taskId"`
	PartNumber int                    `bson:"partNumber"`
	Data       []string               `bson:"data"`
	Percent    float64                `bson:"percent"`
	Status     HashCrackSubtaskStatus `bson:"status"`
	Reason     *string                `bson:"reason,omitempty"`
	CreatedAt  time.Time              `bson:"createdAt"`
	UpdatedAt  time.Time              `bson:"updatedAt"`
}

type HashCrackSubtaskStatus string

const (
	HashCrackSubtaskStatusPending    HashCrackSubtaskStatus = "PENDING"
	HashCrackSubtaskStatusInProgress HashCrackSubtaskStatus = "IN_PROGRESS"
	HashCrackSubtaskStatusSuccess    HashCrackSubtaskStatus = "SUCCESS"
	HashCrackSubtaskStatusError      HashCrackSubtaskStatus = "ERROR"
	HashCrackSubtaskStatusUnknown    HashCrackSubtaskStatus = "UNKNOWN"
)

func (c HashCrackSubtaskStatus) String() string {
	return string(c)
}

func ParseHashCrackSubtaskStatus(s string) HashCrackSubtaskStatus {
	switch s {
	case "PENDING":
		return HashCrackSubtaskStatusPending
	case "IN_PROGRESS":
		return HashCrackSubtaskStatusInProgress
	case "SUCCESS":
		return HashCrackSubtaskStatusSuccess
	case "ERROR":
		return HashCrackSubtaskStatusError
	default:
		return HashCrackSubtaskStatusUnknown
	}
}
