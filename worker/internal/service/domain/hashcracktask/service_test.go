package hashcracktask_test

import (
	"context"
	"crypto/md5"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	mock3 "github.com/stretchr/testify/mock"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher"
	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher/mock"
	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain/hashcracktask"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure"
	mock2 "github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/mock"

	"github.com/stretchr/testify/require"
)

var (
	mockPublisher  *mock.PublisherMock[message.HashCrackTaskResult]
	mockBruteForce *mock2.HashBruteForceMock
	svc            domain.HashCrackTask

	ctx = context.Background()
)

func init() {
	logging.Setup(true)
}

func TestMain(m *testing.M) {
	mockPublisher = new(mock.PublisherMock[message.HashCrackTaskResult])
	mockBruteForce = new(mock2.HashBruteForceMock)
	svc = hashcracktask.NewService(
		log.Logger,
		time.Second,
		mockPublisher,
		mockBruteForce,
	)

	m.Run()
}

func Test_ExecuteTask(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			testCases := []struct {
				name     string
				progress infrastructure.TaskProgress
			}{
				{
					name: "Success",
					progress: infrastructure.TaskProgress{
						Answers: []string{"abc"},
						Percent: 100.0,
						Status:  infrastructure.TaskStatusSuccess,
					},
				},
				{
					name: "Error",
					progress: infrastructure.TaskProgress{
						Answers: []string{"abc"},
						Percent: 100.0,
						Status:  infrastructure.TaskStatusError,
						Reason:  lo.ToPtr("error"),
					},
				},
				{
					name: "Progress",
					progress: infrastructure.TaskProgress{
						Answers: []string{"abc"},
						Percent: 65.0,
						Status:  infrastructure.TaskStatusInProgress,
					},
				},
			}

			for _, tc := range testCases {
				t.Run(
					tc.name, func(t *testing.T) {
						// Arrange
						hash := md5.Sum([]byte("abc"))
						input := &message.HashCrackTaskStarted{
							RequestID:  "123",
							Hash:       string(hash[:]),
							MaxLength:  5,
							PartNumber: 0,
							Alphabet: message.Alphabet{
								Symbols: []string{"a", "b", "c"},
							},
						}
						progressCh := make(chan infrastructure.TaskProgress, 1)
						progressCh <- tc.progress
						close(progressCh)

						mockBruteForce.On(
							"BruteForceMD5", input.Hash, input.Alphabet.Symbols, input.MaxLength, input.PartNumber,
							time.Second,
						).Return(progressCh, nil).Once()
						mockPublisher.On("SendMessage", ctx, mock3.Anything, publisher.Persistent, false, false).
							Run(
								func(args mock3.Arguments) {
									msg, ok := args.Get(1).(*message.HashCrackTaskResult)
									require.True(t, ok)
									require.Equal(t, input.RequestID, msg.RequestID)
									require.Equal(t, input.PartNumber, msg.PartNumber)

									if tc.progress.Status == infrastructure.TaskStatusError {
										require.NotNil(t, msg.Error)
										require.Nil(t, msg.Answer)
										require.Equal(t, string(tc.progress.Status), msg.Status)
										return
									}

									require.Nil(t, msg.Error)
									require.NotNil(t, msg.Answer)
									require.Equal(t, tc.progress.Answers, msg.Answer.Words)
									require.Equal(t, tc.progress.Percent, msg.Answer.Percent)
									require.Equal(t, string(tc.progress.Status), msg.Status)
								},
							).
							Return(nil).Once()

						// Act
						err := svc.ExecuteTask(context.Background(), input)

						// Assert
						require.NoError(t, err)
						mockBruteForce.AssertExpectations(t)
					},
				)
			}
		},
	)

	t.Run(
		"BruteForceError", func(t *testing.T) {
			// Arrange
			hash := md5.Sum([]byte("abc"))
			input := &message.HashCrackTaskStarted{
				RequestID:  "123",
				Hash:       string(hash[:]),
				MaxLength:  5,
				PartNumber: 0,
				Alphabet: message.Alphabet{
					Symbols: []string{"a", "b", "c"},
				},
			}
			expectedError := errors.New("brute force failed")

			mockBruteForce.On(
				"BruteForceMD5", input.Hash, input.Alphabet.Symbols, input.MaxLength, input.PartNumber, time.Second,
			).Return(nil, expectedError).Once()
			mockPublisher.On("SendMessage", ctx, mock3.Anything, publisher.Persistent, false, false).
				Run(
					func(args mock3.Arguments) {
						msg, ok := args.Get(1).(*message.HashCrackTaskResult)
						require.True(t, ok)
						require.Equal(t, input.RequestID, msg.RequestID)
						require.Equal(t, input.PartNumber, msg.PartNumber)
						require.NotNil(t, msg.Error)
						require.Nil(t, msg.Answer)
						require.Equal(t, expectedError.Error(), *msg.Error)
						require.Equal(t, string(infrastructure.TaskStatusError), msg.Status)
					},
				).
				Return(nil).Once()

			// Act
			err := svc.ExecuteTask(context.Background(), input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedError)
			mockBruteForce.AssertExpectations(t)
		},
	)
}
