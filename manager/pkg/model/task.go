package model

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

type HashCrackSubtaskStatusOutput struct {
	Status  string   `json:"status" validate:"required,oneof=PENDING IN_PROGRESS SUCCESS ERROR UNKNOWN"`
	Data    []string `json:"data" validate:"required,min=0,dive,required"`
	Percent float64  `json:"percent" validate:"required,min=0,max=100"`
}
