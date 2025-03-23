package model

type HashCrackTaskInput struct {
	Hash      string `json:"hash" validate:"required"`
	MaxLength int    `json:"maxLength" validate:"min=1,max=6"`
}

type HashCrackTaskIDOutput struct {
	RequestID string `json:"requestId" validate:"required"`
}

type HashCrackTaskStatusOutput struct {
	Status string   `json:"status" validate:"required,oneof=IN_PROGRESS READY PARTIAL_READY ERROR UNKNOWN"`
	Data   []string `json:"data" validate:"required,min=0,dive,required"`
}

type HashCrackTaskWebhookInput struct {
	RequestID  string  `xml:"RequestID" validate:"required"`
	PartNumber int     `xml:"PartNumber"`
	Answer     *Answer `xml:"Answer"`
	Error      *string `xml:"Error"`
}

type Answer struct {
	Words []string `xml:"words" validate:"required,min=0,dive,required"`
}
