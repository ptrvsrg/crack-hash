package hashcrack_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/samber/lo"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
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
	mockTaskRepo            *repomock.HashCrackTaskMock
	mockSubtaskRepo         *repomock.HashCrackSubtaskMock
	mockSplitSvc            *infrasvcmock.TaskSplitMock
	mockTaskWithSubtasksSvc *infrasvcmock.TaskWithSubtasksMock
	mockPublisher           *pubmock.PublisherMock[message.HashCrackTaskStarted]
	cfg                     config.TaskConfig
	service                 domain.HashCrackTask

	ctx = context.Background()
)

func TestMain(m *testing.M) {
	mockTaskRepo = new(repomock.HashCrackTaskMock)
	mockSubtaskRepo = new(repomock.HashCrackSubtaskMock)
	mockSplitSvc = new(infrasvcmock.TaskSplitMock)
	mockTaskWithSubtasksSvc = new(infrasvcmock.TaskWithSubtasksMock)
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
	service = hashcrack.NewService(
		log.Logger, cfg, mockTaskRepo, mockSubtaskRepo, mockSplitSvc, mockTaskWithSubtasksSvc, mockPublisher,
	)

	m.Run()
}

func Test_CreateTask(t *testing.T) {
	t.Run(
		"Split error", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}
			expectedErr := errors.New("split failed")

			mockTaskRepo.On(
				"GetByHashAndMaxLength", ctx, input.Hash, input.MaxLength, false,
			).Return(nil, repository.ErrCrackTaskNotFound).Once()
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

			mockTaskRepo.On(
				"GetByHashAndMaxLength", ctx, input.Hash, input.MaxLength, false,
			).Return(nil, repository.ErrCrackTaskNotFound).Once()
			mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(10, nil).Once()
			mockTaskWithSubtasksSvc.On("CreateTaskWithSubtasks", ctx, mock.Anything).Return(expectedErr).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
			require.Nil(t, output)
		},
	)

	t.Run(
		"Success - already exists", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}

			sameTask := &entity.HashCrackTaskWithSubtasks{
				ObjectID: primitive.NewObjectID(),
			}

			mockTaskRepo.On("GetByHashAndMaxLength", ctx, input.Hash, input.MaxLength, false).
				Return(sameTask, nil).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			// Assert
			require.NoError(t, err)
			require.NotEmpty(t, output.RequestID)
			require.Equal(t, sameTask.ObjectID.Hex(), output.RequestID)
		},
	)

	t.Run(
		"Success - new task", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}

			mockTaskRepo.On(
				"GetByHashAndMaxLength", ctx, input.Hash, input.MaxLength, false,
			).Return(&entity.HashCrackTaskWithSubtasks{}, nil).Once()
			mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(10, nil).Once()
			mockTaskWithSubtasksSvc.On("CreateTaskWithSubtasks", ctx, mock.Anything).Return(nil).Once()
			mockPublisher.On(
				"SendMessage", ctx, mock.Anything, publisher.Persistent, false,
				false,
			).Return(nil).Times(10)
			mockSubtaskRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Times(10)
			mockTaskRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			time.Sleep(time.Second)

			// Assert
			require.NoError(t, err)
			require.NotEmpty(t, output.RequestID)
		},
	)
}

func Test_GetTaskStatus(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			// Arrange
			objID := primitive.NewObjectID()
			taskID := objID.Hex()
			task := &entity.HashCrackTaskWithSubtasks{
				ObjectID: objID,
				Status:   entity.HashCrackTaskStatusReady,
				Subtasks: []*entity.HashCrackSubtask{
					{
						ObjectID: objID,
						Status:   entity.HashCrackSubtaskStatusSuccess,
						Data:     []string{"data"},
					},
				},
			}

			mockTaskRepo.On("Get", ctx, mock.Anything, true).Run(
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
			require.Len(t, output.Data, 1)
			require.Equal(t, task.Subtasks[0].Data[0], output.Data[0])
			require.Len(t, output.Subtasks, 1)
			require.Equal(t, entity.HashCrackSubtaskStatusSuccess.String(), output.Subtasks[0].Status)
			require.Equal(t, task.Subtasks[0].Data, output.Subtasks[0].Data)
		},
	)

	t.Run(
		"Task not found", func(t *testing.T) {
			// Arrange
			objID := primitive.NewObjectID()
			taskID := objID.Hex()

			mockTaskRepo.On("Get", ctx, mock.Anything, true).Run(
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
		"WithTransaction error", func(t *testing.T) {
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

			expectedError := errors.New("error")

			mockTaskRepo.EXPECT().WithTransaction(ctx, mock.Anything).Return(nil, expectedError).Once()

			// Act
			err := service.SaveResultSubtask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedError)
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

			mockTaskRepo.EXPECT().WithTransaction(ctx, mock.Anything).RunAndReturn(
				func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
					return fn(ctx)
				}).Once()
			mockTaskRepo.On("Get", mock.Anything, mock.Anything, true).Run(
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
							PartNumber: 0,
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
						task := &entity.HashCrackTaskWithSubtasks{
							ObjectID:  objID,
							PartCount: 2,
						}

						// Subtask
						switch {
						case c.SubtaskStatus == entity.HashCrackSubtaskStatusSuccess &&
							c.TaskStatus == entity.HashCrackTaskStatusReady ||
							c.SubtaskStatus == entity.HashCrackSubtaskStatusError &&
								c.TaskStatus == entity.HashCrackTaskStatusPartialReady:

							task.Subtasks = []*entity.HashCrackSubtask{
								{
									PartNumber: 0,
									Status:     entity.HashCrackSubtaskStatusSuccess,
								},
								{
									PartNumber: 1,
									Status:     entity.HashCrackSubtaskStatusInProgress,
								},
							}

						default:

							task.Subtasks = []*entity.HashCrackSubtask{
								{
									PartNumber: 0,
									Status:     entity.HashCrackSubtaskStatusError,
								},
								{
									PartNumber: 1,
									Status:     entity.HashCrackSubtaskStatusInProgress,
								},
							}

						}

						mockTaskRepo.EXPECT().WithTransaction(ctx, mock.Anything).RunAndReturn(
							func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
								return fn(ctx)
							}).Once()
						mockTaskRepo.On("Get", mock.Anything, mock.Anything, true).Run(
							func(args mock.Arguments) {
								objectID, ok := args.Get(1).(primitive.ObjectID)
								assert.True(t, ok)
								assert.Equal(t, objID.Hex(), objectID.Hex())
							},
						).Return(task, nil).Once()

						mockSubtaskRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
						call := mockTaskRepo.On("Update", mock.Anything, mock.Anything).Run(
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

			task := &entity.HashCrackTaskWithSubtasks{
				ObjectID:  objID,
				PartCount: 2,
				Subtasks: []*entity.HashCrackSubtask{
					{
						PartNumber: 0,
						Status:     entity.HashCrackSubtaskStatusInProgress,
					},
					{
						PartNumber: 1,
						Status:     entity.HashCrackSubtaskStatusInProgress,
					},
				},
			}

			mockTaskRepo.EXPECT().WithTransaction(ctx, mock.Anything).RunAndReturn(
				func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
					return fn(ctx)
				}).Once()
			mockTaskRepo.On("Get", mock.Anything, mock.Anything, true).Run(
				func(args mock.Arguments) {
					objectID, ok := args.Get(1).(primitive.ObjectID)
					assert.True(t, ok)
					assert.Equal(t, objID.Hex(), objectID.Hex())
				},
			).Return(task, nil).Once()
			mockTaskRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			mockSubtaskRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

			// Act
			err := service.SaveResultSubtask(ctx, input)

			// Assert
			require.NoError(t, err)
		},
	)
}

func Test_FinishTasks(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			// Arrange
			tasks := []*entity.HashCrackTaskWithSubtasks{
				{
					ObjectID: primitive.NewObjectID(),
					Status:   "PENDING",
				},
			}

			mockTaskRepo.On("GetAllFinished", ctx, true).Return(tasks, nil).Once()
			mockTaskWithSubtasksSvc.On("UpdateTaskWithSubtasks", ctx, mock.Anything).Return(nil).Once()

			// Act
			err := service.FinishTimeoutTasks(ctx)

			// Assert
			require.NoError(t, err)
		},
	)

	t.Run(
		"No tasks", func(t *testing.T) {
			// Arrange
			mockTaskRepo.On("GetAllFinished", ctx, true).Return([]*entity.HashCrackTaskWithSubtasks{}, nil).Once()

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
			mockTaskRepo.On("GetAllFinished", ctx, true).Return(nil, expectedError).Once()

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
			tasks := []*entity.HashCrackTaskWithSubtasks{
				{
					ObjectID: primitive.NewObjectID(),
					Status:   "PENDING",
				},
			}

			expectedError := errors.New("update failed")

			mockTaskRepo.On("GetAllFinished", ctx, true).Return(tasks, nil).Once()
			mockTaskWithSubtasksSvc.On("UpdateTaskWithSubtasks", ctx, mock.Anything).Return(expectedError).Once()

			// Act
			err := service.FinishTimeoutTasks(ctx)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedError)
		},
	)
}

func Test_DeleteExpiredTasks(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			// Arrange
			tasks := []*entity.HashCrackTaskWithSubtasks{
				{
					ObjectID: primitive.NewObjectID(),
					Status:   "PENDING",
				},
			}

			mockTaskRepo.On("GetAllExpired", ctx, mock.Anything, true).Return(tasks, nil).Once()
			mockTaskWithSubtasksSvc.On("DeleteTasksWithSubtasks", ctx, tasks).Return(nil).Once()

			// Act
			err := service.DeleteExpiredTasks(ctx)

			// Assert
			require.NoError(t, err)
		},
	)

	t.Run(
		"GetAllExpired error", func(t *testing.T) {
			// Arrange
			expectedErr := errors.New("repo error")
			mockTaskRepo.On("GetAllExpired", ctx, mock.Anything, true).Return(nil, expectedErr).Once()

			// Act
			err := service.DeleteExpiredTasks(ctx)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
		},
	)

	t.Run(
		"No tasks", func(t *testing.T) {
			// Arrange
			tasks := []*entity.HashCrackTaskWithSubtasks{}

			mockTaskRepo.On("GetAllExpired", ctx, mock.Anything, true).Return(tasks, nil).Once()
			mockTaskWithSubtasksSvc.On("DeleteTaskWithSubtasks", ctx, mock.Anything).Return(nil).Once()

			// Act
			err := service.DeleteExpiredTasks(ctx)

			// Assert
			require.NoError(t, err)
		},
	)

	t.Run(
		"DeleteTaskWithSubtasks error", func(t *testing.T) {
			// Arrange
			tasks := []*entity.HashCrackTaskWithSubtasks{
				{
					ObjectID: primitive.NewObjectID(),
					Status:   "PENDING",
				},
			}

			expectedErr := errors.New("repo error")

			mockTaskRepo.On("GetAllExpired", ctx, mock.Anything, true).Return(tasks, nil).Once()
			mockTaskWithSubtasksSvc.On("DeleteTasksWithSubtasks", ctx, tasks).Return(expectedErr).Once()

			// Act
			err := service.DeleteExpiredTasks(ctx)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
		},
	)
}

func Test_ExecutePendingSubtasks(t *testing.T) {
	t.Run(
		"GetAllByStatus error", func(t *testing.T) {
			// Arrange
			expectedErr := errors.New("repo error")

			mockSubtaskRepo.EXPECT().GetAllByStatus(ctx, entity.HashCrackSubtaskStatusPending).
				Return(nil, expectedErr).Once()

			// Act
			err := service.ExecutePendingSubtasks(ctx)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
		},
	)

	t.Run(
		"Get task error", func(t *testing.T) {
			// Arrange
			subtasks := []*entity.HashCrackSubtask{
				{
					ObjectID: primitive.NewObjectID(),
					Status:   entity.HashCrackSubtaskStatusPending,
					TaskID:   primitive.NewObjectID(),
				},
			}

			expectedErr := errors.New("repo error")

			mockSubtaskRepo.EXPECT().GetAllByStatus(ctx, entity.HashCrackSubtaskStatusPending).
				Return(subtasks, nil).Once()
			mockTaskRepo.EXPECT().Get(ctx, subtasks[0].TaskID, false).
				Return(nil, expectedErr).Once()

			// Act
			err := service.ExecutePendingSubtasks(ctx)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
		},
	)

	t.Run(
		"Success - No subtasks", func(t *testing.T) {
			// Arrange
			mockSubtaskRepo.EXPECT().GetAllByStatus(ctx, entity.HashCrackSubtaskStatusPending).
				Return([]*entity.HashCrackSubtask{}, nil).Once()

			// Act
			err := service.ExecutePendingSubtasks(ctx)

			// Assert
			require.NoError(t, err)
		},
	)

	t.Run(
		"Success - Has subtasks", func(t *testing.T) {
			// Arrange
			subtasks := []*entity.HashCrackSubtask{
				{
					ObjectID: primitive.NewObjectID(),
					Status:   entity.HashCrackSubtaskStatusPending,
					TaskID:   primitive.NewObjectID(),
				},
			}
			task := &entity.HashCrackTaskWithSubtasks{
				ObjectID: subtasks[0].TaskID,
				Status:   entity.HashCrackTaskStatusInProgress,
			}

			mockSubtaskRepo.EXPECT().GetAllByStatus(ctx, entity.HashCrackSubtaskStatusPending).
				Return(subtasks, nil).Once()
			mockTaskRepo.EXPECT().Get(ctx, subtasks[0].TaskID, false).
				Return(task, nil).Once()
			mockPublisher.EXPECT().SendMessage(ctx, mock.Anything, publisher.Persistent, false, false).
				Return(nil).Times(1)
			mockTaskRepo.EXPECT().Update(ctx, task.ToHashCrackTask()).Return(nil).Once()
			mockSubtaskRepo.EXPECT().Update(ctx, subtasks[0]).Return(nil).Times(1)

			// Act
			err := service.ExecutePendingSubtasks(ctx)

			// Assert
			require.NoError(t, err)
		},
	)
}
