package hashcrack_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"

	"github.com/ptrvsrg/crack-hash/commonlib/http/client"
	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	repomock "github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository/mock"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain/hashcrack"
	infrasvcmock "github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/mock"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

func init() {
	logging.Setup(true)
}

var (
	mockRepo     *repomock.HashCrackTaskMock
	mockSplitSvc *infrasvcmock.TaskSplitMock
	httpClient   *resty.Client
	cfg          config.TaskConfig
	service      domain.HashCrackTask

	ctx = context.Background()
)

func TestMain(m *testing.M) {
	// HTTP server
	router := gin.New()
	router.POST(
		"/internal/api/worker/hash/crack/task", func(c *gin.Context) {
			log.Debug().Msg("handle hash crack task")
			c.JSON(http.StatusOK, nil)
		},
	)

	server := httptest.NewServer(router)
	defer server.Close()

	// HTTP httpClient
	var err error
	httpClient, err = client.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create HTTP httpClient")
	}
	defer func(httpClient *resty.Client) {
		err := httpClient.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close HTTP httpClient")
		}
	}(httpClient)

	httpClient.SetBaseURL(server.URL)

	// Services
	mockRepo = new(repomock.HashCrackTaskMock)
	mockSplitSvc = new(infrasvcmock.TaskSplitMock)
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
	service = hashcrack.NewService(cfg, httpClient, mockRepo, mockSplitSvc)

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
					ID:        uuid.New().String(),
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
			require.Equal(t, sameTasks[0].ID, output.RequestID)
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
			mockRepo.On("CountByStatus", ctx, entity.HashCrackTaskStatusInProgress).Return(0, nil).Once()
			mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(10, nil).Once()
			mockRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
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
		"Count IN_PROGRESS error", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}
			expectedErr := errors.New("count failed")

			mockRepo.On(
				"GetAllByHashAndMaxLength", ctx, input.Hash,
				input.MaxLength,
			).Return([]*entity.HashCrackTask{}, nil).Once()
			mockRepo.On("CountByStatus", ctx, entity.HashCrackTaskStatusInProgress).Return(0, expectedErr).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, expectedErr)
			require.Nil(t, output)
		},
	)

	t.Run(
		"Too many tasks", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskInput{
				MaxLength: 5,
				Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
			}

			mockRepo.On(
				"GetAllByHashAndMaxLength", ctx, input.Hash,
				input.MaxLength,
			).Return([]*entity.HashCrackTask{}, nil).Once()
			mockRepo.On("CountByStatus", ctx, entity.HashCrackTaskStatusInProgress).Return(cfg.Limit, nil).Once()

			// Act
			output, err := service.CreateTask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, domain.ErrTooManyTasks)
			require.Nil(t, output)
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
			taskID := "123"
			task := &entity.HashCrackTask{
				ID:     taskID,
				Status: "READY",
			}

			mockRepo.On("Get", mock.Anything, taskID).Return(task, nil).Once()

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
			taskID := "123"
			mockRepo.On("Get", mock.Anything, taskID).Return(nil, repository.ErrCrackTaskNotFound).Once()

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
			}

			for _, c := range cases {
				t.Run(
					c.Name, func(t *testing.T) {
						// Arrange

						// Webhook input
						input := &model.HashCrackTaskWebhookInput{
							RequestID:  "123",
							PartNumber: 1,
						}

						if c.SubtaskStatus == entity.HashCrackSubtaskStatusSuccess {
							input.Answer = &model.Answer{
								Words: []string{"word1", "word2"},
							}
						} else {
							input.Error = lo.ToPtr("error")
						}

						// Task
						task := &entity.HashCrackTask{
							ID:        input.RequestID,
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

						mockRepo.On("Get", mock.Anything, input.RequestID).Return(task, nil).Once()
						mockRepo.On("Update", mock.Anything, mock.Anything).Run(
							func(args mock.Arguments) {
								task, ok := args.Get(1).(*entity.HashCrackTask)
								assert.True(t, ok)

								if task.Status == "" {
									return
								}
								assert.Equal(t, c.TaskStatus, task.Status)
							},
						).Return(nil).Twice()

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
			input := &model.HashCrackTaskWebhookInput{
				RequestID:  "123",
				PartNumber: 0,
				Answer: &model.Answer{
					Words: []string{"word1", "word2"},
				},
			}

			task := &entity.HashCrackTask{
				ID:        input.RequestID,
				PartCount: 2,
				Subtasks:  make(map[int]*entity.HashCrackSubtask),
			}

			mockRepo.On("Get", mock.Anything, input.RequestID).Return(task, nil).Once()
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
			input := &model.HashCrackTaskWebhookInput{
				RequestID: "123",
				Answer: &model.Answer{
					Words: []string{"word1", "word2"},
				},
			}

			mockRepo.On("Get", mock.Anything, input.RequestID).Return(nil, repository.ErrCrackTaskNotFound).Once()

			// Act
			err := service.SaveResultSubtask(ctx, input)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, repository.ErrCrackTaskNotFound)
		},
	)

	t.Run(
		"Update error", func(t *testing.T) {
			// Arrange
			input := &model.HashCrackTaskWebhookInput{
				RequestID:  "123",
				PartNumber: 0,
				Answer: &model.Answer{
					Words: []string{"word1", "word2"},
				},
			}

			task := &entity.HashCrackTask{
				ID:        input.RequestID,
				PartCount: 1,
				Subtasks:  make(map[int]*entity.HashCrackSubtask),
			}

			expectedError := errors.New("update failed")

			mockRepo.On("Get", mock.Anything, input.RequestID).Return(task, nil).Once()
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
					ID:     "123",
					Status: "PENDING",
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
					ID:     "123",
					Status: "PENDING",
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
