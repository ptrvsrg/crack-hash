package infrastructure

import "time"

const (
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusSuccess    TaskStatus = "SUCCESS"
	TaskStatusError      TaskStatus = "ERROR"
)

type (
	TaskStatus string

	TaskProgress struct {
		Answers []string
		Percent float64
		Status  TaskStatus
		Reason  *string
	}
)

type HashBruteForce interface {
	BruteForceMD5(
		hash string, alphabet []string, maxLength, partNumber int, progressPeriod time.Duration,
	) (<-chan TaskProgress, error)
}

type Services struct {
	HashBruteForce HashBruteForce
}
