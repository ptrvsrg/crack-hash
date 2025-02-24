package entity

type HashCrackSubtask struct {
	PartNumber int                    `bson:"partNumber"`
	Data       []string               `bson:"data"`
	Percent    float64                `bson:"percent"`
	Status     HashCrackSubtaskStatus `bson:"status"`
	Reason     *string                `bson:"reason,omitempty"`
}

type HashCrackSubtaskStatus string

const (
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
