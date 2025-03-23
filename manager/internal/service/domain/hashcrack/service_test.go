package hashcrack_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher"
	pubmock "github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher/mock"
	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	repomock "github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository/mock"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain/hashcrack"
	infrasvcmock "github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/mock"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

func init() {
	logging.Setup(true)
}

var (
	mockRepo      *repomock.HashCrackTaskMock
	mockSplitSvc  *infrasvcmock.TaskSplitMock
	mockPublisher *pubmock.PublisherMock[message.HashCrackTaskStarted]
	cfg           config.TaskConfig
	service       domain.HashCrackTask

	ctx = context.Background()
)

func TestMain(m *testing.M) {
	mockRepo = new(repomock.HashCrackTaskMock)
	mockSplitSvc = new(infrasvcmock.TaskSplitMock)
	mockPublisher = new(pubmock.PublisherMock[message.HashCrackTaskStarted])
	cfg = config.TaskConfig{
		Split: config.TaskSplitConfig{
			Strategy:  "chunkBased",
			ChunkSize: 10,
		},
		Timeout:     time.Hour,
		Limit:       10,
		MaxAge:      time.Hour * 24,
		FinishDelay: time.Minute,
	}
	service = hashcrack.NewService(log.Logger, cfg, mockRepo, mockSplitSvc, mockPublisher)

	m.Run()
}

func Test_CreateTask(t *testing.T) {
	t.Run(
		"Success - already exists", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}

			sameTasks := []*entity.HashCrackTask{
				{
					ObjectID:  primitive.NewObjectID(),
					Hash:      input.Hash,
					MaxLength: input.MaxLength,
				},
			}

			mockRepo.On("GetAllByHashAndMaxLength", ctx, input.Hash, input.MaxLength).Return(sameTasks, nil).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			// Assert
			require.NoError(t, err)
			require.NotEmpty(t, output.RequestID)
			require.Equal(t, sameTasks[0].ObjectID.Hex(), output.RequestID)
		},
	)

	t.Run(
		"Success - new task", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}

			mockRepo.On(
				"GetAllByHashAndMaxLength", ctx, input.Hash,
				input.MaxLength,
			).Return([]*entity.HashCrackTask{}, nil).Once()
			mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(10, nil).Once()
			mockRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
			mockPublisher.On(
				"SendMessage", ctx, mock.Anything, publisher.Persistent, false,
				false,
			).Return(nil).Times(10)
			mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			time.Sleep(time.Second)

			// Assert
			require.NoError(t, err)
			require.NotEmpty(t, output.RequestID)
		},
	)

	t.Run(
		"Split error", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}
			expectedErr := errors.New("split failed")

			mockRepo.On(
				"GetAllByHashAndMaxLength", ctx, input.Hash,
				input.MaxLength,
			).Return([]*entity.HashCrackTask{}, nil).Once()
			mockRepo.On("CountByStatus", ctx, entity.HashCrackTaskStatusInProgress).Return(0, nil).Once()
			mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(0, expectedErr).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
			require.Nil(t, output)
		},
	)

	t.Run(
		"Create error", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}
			expectedErr := errors.New("create failed")

			mockRepo.On(
				"GetAllByHashAndMaxLength", ctx, input.Hash,
				input.MaxLength,
			).Return([]*entity.HashCrackTask{}, nil).Once()
			mockRepo.On("CountByStatus", ctx, entity.HashCrackTaskStatusInProgress).Return(0, nil).Once()
			mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(10, nil).Once()
			mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedErr).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
			require.Nil(t, output)
		},
	)
}

func Test_GetTaskStatus(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			// Arrange
			objID := primitive.NewObjectID()
			taskID := objID.Hex()
			task := &entity.HashCrackTask{
				ObjectID: objID,
				Status:   "READY",
			}

			mockRepo.On("Get", mock.Anything, mock.Anything).Run(
				func(args mock.Arguments) {
					objectID, ok := args.Get(1).(primitive.ObjectID)
					assert.True(t, ok)
					assert.Equal(t, taskID, objectID.Hex())
				},
			).Return(task, nil).Once()

			// Act
			output, err := service.GetTaskStatus(ctx, taskID)

			// Assert
			require.NoError(t, err)
			require.Equal(t, entity.HashCrackTaskStatusReady.String(), output.Status)
		},
	)

	t.Run(
		"Task not found", func(t *testing.T) {
			// Arrange
			objID := primitive.NewObjectID()
			taskID := objID.Hex()

			mockRepo.On("Get", mock.Anything, mock.Anything).Run(
				func(args mock.Arguments) {
					objectID, ok := args.Get(1).(primitive.ObjectID)
					assert.True(t, ok)
					assert.Equal(t, taskID, objectID.Hex())
				},
			).Return(nil, repository.ErrCrackTaskNotFound).Once()

			// Act
			output, err := service.GetTaskStatus(ctx, taskID)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, domain.ErrTaskNotFound)
			require.Nil(t, output)
		},
	)
}

func Test_SaveResultTask(t *testing.T) {
	t.Run(
		"Success - Finish task", func(t *testing.T) {
			cases := []struct {
				Name          string
				SubtaskStatus entity.HashCrackSubtaskStatus
				TaskStatus    entity.HashCrackTaskStatus
			}{
				{
					"SUCCESS webhook - READY status",
					entity.HashCrackSubtaskStatusSuccess,
					entity.HashCrackTaskStatusReady,
				},
				{
					"SUCCESS webhook - PARTIAL_READY status",
					entity.HashCrackSubtaskStatusSuccess,
					entity.HashCrackTaskStatusPartialReady,
				},
				{
					"ERROR webhook - PARTIAL_READY status",
					entity.HashCrackSubtaskStatusError,
					entity.HashCrackTaskStatusPartialReady,
				},
				{
					"ERROR webhook - ERROR status",
					entity.HashCrackSubtaskStatusError,
					entity.HashCrackTaskStatusError,
				},
				{
					"IN_PROGRESS webhook - IN_PROGRESS status",
					entity.HashCrackSubtaskStatusInProgress,
					entity.HashCrackTaskStatusInProgress,
				},
			}

			for _, c := range cases {
				t.Run(
					c.Name, func(t *testing.T) {
						// Arrange
						objID := primitive.NewObjectID()
						input := &message.HashCrackTaskResult{
							RequestID:  objID.Hex(),
							PartNumber: 1,
							Status:     c.SubtaskStatus.String(),
						}

						switch {
						case c.SubtaskStatus == entity.HashCrackSubtaskStatusSuccess:
							input.Answer = &message.Answer{
								Words:   []string{"word1", "word2"},
								Percent: 100.0,
							}

						case c.SubtaskStatus == entity.HashCrackSubtaskStatusInProgress:
							input.Answer = &message.Answer{
								Words:   []string{"word1", "word2"},
								Percent: 50.0,
							}

						case c.SubtaskStatus == entity.HashCrackSubtaskStatusError:
							input.Error = lo.ToPtr("error")
						}

						// Task
						task := &entity.HashCrackTask{
							ObjectID:  objID,
							PartCount: 2,
						}

						// Subtask
						switch {
						case c.SubtaskStatus == entity.HashCrackSubtaskStatusSuccess &&
							c.TaskStatus == entity.HashCrackTaskStatusReady ||
							c.SubtaskStatus == entity.HashCrackSubtaskStatusError &&
								c.TaskStatus == entity.HashCrackTaskStatusPartialReady:

							task.Subtasks = map[int]*entity.HashCrackSubtask{
								0: {
									PartNumber: 0,
									Status:     entity.HashCrackSubtaskStatusSuccess,
								},
							}

						default:

							task.Subtasks = map[int]*entity.HashCrackSubtask{
								0: {
									PartNumber: 0,
									Status:     entity.HashCrackSubtaskStatusError,
								},
							}

						}

						mockRepo.On("Get", mock.Anything, mock.Anything).Run(
							func(args mock.Arguments) {
								objectID, ok := args.Get(1).(primitive.ObjectID)
								assert.True(t, ok)
								assert.Equal(t, objID.Hex(), objectID.Hex())
							},
						).Return(task, nil).Once()

						call := mockRepo.On("Update", mock.Anything, mock.Anything).Run(
							func(args mock.Arguments) {
								task, ok := args.Get(1).(*entity.HashCrackTask)
								assert.True(t, ok)

								if task.Status == "" {
									return
								}
								assert.Equal(t, c.TaskStatus, task.Status)
							},
						).Return(nil)

						if c.SubtaskStatus == entity.HashCrackSubtaskStatusInProgress {
							call.Once()
						} else {
							call.Twice()
						}

						// Act
						err := service.SaveResultSubtask(ctx, input)

						// Assert
						require.NoError(t, err)
					},
				)
			}
		},
	)

	t.Run(
		"Success - Not finish task", func(t *testing.T) {
			// Arrange
			objID := primitive.NewObjectID()
			input := &message.HashCrackTaskResult{
				RequestID:  objID.Hex(),
				PartNumber: 0,
				Answer: &message.Answer{
					Words:   []string{"word1", "word2"},
					Percent: 100.0,
				},
				Status: entity.HashCrackSubtaskStatusSuccess.String(),
			}

			task := &entity.HashCrackTask{
				ObjectID:  objID,
				PartCount: 2,
				Subtasks:  make(map[int]*entity.HashCrackSubtask),
			}

			mockRepo.On("Get", mock.Anything, mock.Anything).Run(
				func(args mock.Arguments) {
					objectID, ok := args.Get(1).(primitive.ObjectID)
					assert.True(t, ok)
					assert.Equal(t, objID.Hex(), objectID.Hex())
				},
			).Return(task, nil).Once()
			mockRepo.On("Update", mock.Anything, mock.Anything).Run(
				func(args mock.Arguments) {
					task, ok := args.Get(1).(*entity.HashCrackTask)
					assert.True(t, ok)
					assert.Len(t, task.Subtasks, 1)
				},
			).Return(nil).Once()

			// Act
			err := service.SaveResultSubtask(ctx, input)

			// Assert
			require.NoError(t, err)
		},
	)

	t.Run(
		"Task not found", func(t *testing.T) {
			// Arrange
			objID := primitive.NewObjectID()
			input := &message.HashCrackTaskResult{
				RequestID: objID.Hex(),
				Answer: &message.Answer{
					Words:   []string{"word1", "word2"},
					Percent: 100.0,
				},
				Status: entity.HashCrackSubtaskStatusSuccess.String(),
			}

			mockRepo.On("Get", mock.Anything, mock.Anything).Run(
				func(args mock.Arguments) {
					objectID, ok := args.Get(1).(primitive.ObjectID)
					assert.True(t, ok)
					assert.Equal(t, objID.Hex(), objectID.Hex())
				},
			).Return(nil, repository.ErrCrackTaskNotFound).Once()

			// Act
			err := service.SaveResultSubtask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, domain.ErrTaskNotFound)
		},
	)

	t.Run(
		"Update error", func(t *testing.T) {
			// Arrange
			objID := primitive.NewObjectID()
			input := &message.HashCrackTaskResult{
				RequestID:  objID.Hex(),
				PartNumber: 0,
				Answer: &message.Answer{
					Words:   []string{"word1", "word2"},
					Percent: 100.0,
				},
				Status: entity.HashCrackSubtaskStatusSuccess.String(),
			}

			task := &entity.HashCrackTask{
				ObjectID:  objID,
				PartCount: 1,
				Subtasks:  make(map[int]*entity.HashCrackSubtask),
			}

			expectedError := errors.New("update failed")

			mockRepo.On("Get", mock.Anything, mock.Anything).Run(
				func(args mock.Arguments) {
					objectID, ok := args.Get(1).(primitive.ObjectID)
					assert.True(t, ok)
					assert.Equal(t, objID.Hex(), objectID.Hex())
				},
			).Return(task, nil).Once()
			mockRepo.On("Update", mock.Anything, mock.Anything).Return(expectedError).Once()

			// Act
			err := service.SaveResultSubtask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedError)
		},
	)
}

func TestFinishTasks(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			// Arrange
			tasks := []*entity.HashCrackTask{
				{
					ObjectID: primitive.NewObjectID(),
					Status:   "PENDING",
				},
			}

			mockRepo.On("GetAllFinished", mock.Anything).Return(tasks, nil).Once()
			mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

			// Act
			err := service.FinishTimeoutTasks(ctx)

			// Assert
			require.NoError(t, err)
		},
	)

	t.Run(
		"No tasks", func(t *testing.T) {
			// Arrange
			mockRepo.On("GetAllFinished", mock.Anything).Return([]*entity.HashCrackTask{}, nil).Once()

			// Act
			err := service.FinishTimeoutTasks(ctx)

			// Assert
			require.NoError(t, err)
		},
	)

	t.Run(
		"GetAllFinished error", func(t *testing.T) {
			// Arrange
			expectedError := errors.New("get all finished failed")
			mockRepo.On("GetAllFinished", mock.Anything).Return(nil, expectedError).Once()

			// Act
			err := service.FinishTimeoutTasks(ctx)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedError)
		},
	)

	t.Run(
		"Update error", func(t *testing.T) {
			// Arrange
			tasks := []*entity.HashCrackTask{
				{
					ObjectID: primitive.NewObjectID(),
					Status:   "PENDING",
				},
			}

			expectedError := errors.New("update failed")

			mockRepo.On("GetAllFinished", mock.Anything).Return(tasks, nil).Once()
			mockRepo.On("Update", mock.Anything, mock.Anything).Return(expectedError).Once()

			// Act
			err := service.FinishTimeoutTasks(ctx)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedError)
		},
	)
}

func TestDeleteExpiredTasks(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			// Arrange
			mockRepo.On("DeleteAllExpired", mock.Anything, mock.Anything).Return(nil).Once()

			// Act
			err := service.DeleteExpiredTasks(ctx)

			// Assert
			require.NoError(t, err)
		},
	)

	t.Run(
		"DeleteAllExpired error", func(t *testing.T) {
			// Arrange
			expectedErr := errors.New("repo error")
			mockRepo.On("DeleteAllExpired", mock.Anything, mock.Anything).Return(expectedErr).Once()

			// Act
			err := service.DeleteExpiredTasks(ctx)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
		},
	)
}
