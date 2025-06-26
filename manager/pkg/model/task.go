package model

import "time"

type HashCrackTaskInput struct {
	Hash      string `json:"hash" validate:"required"`
	MaxLength int    `json:"maxLength" validate:"required,min=1,max=6"`
}

type HashCrackTaskIDOutput struct {
	RequestID string `json:"requestId" validate:"required"`
}

type HashCrackTaskStatusOutput struct {
	Status   string                         `json:"status" validate:"required,oneof=PENDING IN_PROGRESS READY PARTIAL_READY ERROR UNKNOWN"`
	Data     []string                       `json:"data" validate:"required,min=0,dive,required"`
	Percent  float64                        `json:"percent" validate:"required,min=0,max=100"`
	Subtasks []HashCrackSubtaskStatusOutput `json:"subtasks" validate:"required,min=0,dive"`
}

type HashCrackTaskMetadataInput struct {
	Limit  int `form:"limit,default=10" validate:"required,min=0"`
	Offset int `form:"offset,default=0" validate:"required,min=0"`
}

type HashCrackTaskMetadataOutput struct {
	RequestID string    `json:"requestId" validate:"required"`
	Hash      string    `json:"hash" validate:"required"`
	MaxLength int       `json:"maxLength" validate:"required,min=1,max=6"`
	CreatedAt time.Time `json:"createdAt" validate:"required"`
}

type HashCrackTaskMetadatasOutput struct {
	Count int64                          `json:"count" validate:"required,min=0"`
	Tasks []*HashCrackTaskMetadataOutput `json:"tasks" validate:"required,min=0,dive"`
}

type HashCrackSubtaskStatusOutput struct {
	Status  string   `json:"status" validate:"required,oneof=PENDING IN_PROGRESS SUCCESS ERROR UNKNOWN"`
	Data    []string `json:"data" validate:"required,min=0,dive,required"`
	Percent float64  `json:"percent" validate:"required,min=0,max=100"`
}
