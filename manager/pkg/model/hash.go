package model

type HashCrackTaskInput struct {
	Hash      string `json:"hash" validate:"required"`
	MaxLength int    `json:"maxLength" validate:"min=1,max=6"`
}

type HashCrackTaskIDOutput struct {
	RequestID string `json:"requestId"`
}

type HashCrackTaskStatusOutput struct {
	Status HashCrackTaskStatus `json:"status"`
	Data   []string            `json:"data"`
}

type HashCrackTaskStatus string

const (
	HashCrackStatusInProgress HashCrackTaskStatus = "IN_PROGRESS"
	HashCrackStatusReady      HashCrackTaskStatus = "READY"
	HashCrackStatusError      HashCrackTaskStatus = "ERROR"
	HashCrackStatusUnknown    HashCrackTaskStatus = "UNKNOWN"
)

func (c HashCrackTaskStatus) String() string {
	return string(c)
}

func ParseHashCrackTaskStatus(s string) HashCrackTaskStatus {
	switch s {
	case "IN_PROGRESS":
		return HashCrackStatusInProgress
	case "READY":
		return HashCrackStatusReady
	case "ERROR":
		return HashCrackStatusError
	default:
		return HashCrackStatusUnknown
	}
}

type HashCrackTaskWebhookInput struct {
	RequestID  string `xml:"RequestId" binding:"required"`
	PartNumber int    `xml:"PartNumber"`
	Answer     struct {
		Words []string `xml:"words"`
	} `xml:"Answer" binding:"required"`
}
