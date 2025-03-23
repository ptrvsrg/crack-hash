package entity

type HashCrackSubtask struct {
	PartNumber int
	Data       []string
	Status     HashCrackSubtaskStatus
	Reason     *string
}

type HashCrackSubtaskStatus string

const (
	HashCrackSubtaskStatusSuccess HashCrackSubtaskStatus = "SUCCESS"
	HashCrackSubtaskStatusError   HashCrackSubtaskStatus = "ERROR"
	HashCrackSubtaskStatusUnknown HashCrackSubtaskStatus = "UNKNOWN"
)

func (c HashCrackSubtaskStatus) String() string {
	return string(c)
}

func ParseHashCrackSubtaskStatus(s string) HashCrackSubtaskStatus {
	switch s {
	case "SUCCESS":
		return HashCrackSubtaskStatusSuccess
	case "ERROR":
		return HashCrackSubtaskStatusError
	default:
		return HashCrackSubtaskStatusUnknown
	}
}
