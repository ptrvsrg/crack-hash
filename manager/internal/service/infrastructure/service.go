package infrastructure

import (
	"context"
	"errors"
)

var (
	ErrInvalidAlphabetLength = errors.New("invalid alphabet length")
	ErrInvalidWordMaxLength  = errors.New("invalid word max length")
)

type TaskSplit interface {
	Split(ctx context.Context, wordMaxLength, alphabetLength int) (int, error)
}

type Services struct {
	TaskSplit TaskSplit
}
