package hashcrack_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/logging"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	repomock "github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository/mock"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain/hashcrack"
	infrasvcmock "github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/mock"
	"github.com/ptrvsrg/crack-hash/manager/pkg/model"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"resty.dev/v3"
	"strings"
	"testing"
	"time"
)

func init() {
	logging.Setup(config.EnvDev)
}

var (
	mockRepo     *repomock.HashCrackTaskMock
	mockSplitSvc *infrasvcmock.TaskSplitMock
	client       *resty.Client
	cfg          config.WorkerConfig
	service      domain.HashCrackTask

	ctx = context.Background()
)

func TestMain(m *testing.M) {
	router := gin.New()
	router.POST("/internal/api/worker/hash/crack/task", func(c *gin.Context) {
		log.Debug().Msg("handle hash crack task")
		c.JSON(http.StatusOK, nil)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "", 1)

	mockRepo = new(repomock.HashCrackTaskMock)
	mockSplitSvc = new(infrasvcmock.TaskSplitMock)
	client = resty.New()
	cfg = config.WorkerConfig{Address: url}
	service = hashcrack.NewService(cfg, client, mockRepo, mockSplitSvc)

	m.Run()
}

func Test_CreateTask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		input := &model.HashCrackTaskInput{
			MaxLength: 5,
			Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
		}

		mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(10, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		// Act
		output, err := service.CreateTask(ctx, input)

		time.Sleep(time.Second)

		// Assert
		require.NoError(t, err)
		require.NotEmpty(t, output.RequestID)

		mockRepo.AssertExpectations(t)
		mockSplitSvc.AssertExpectations(t)
	})

	t.Run("Split error", func(t *testing.T) {
		// Arrange
		input := &model.HashCrackTaskInput{
			MaxLength: 5,
			Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
		}
		expectedErr := errors.New("split failed")

		mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(0, expectedErr).Once()

		// Act
		output, err := service.CreateTask(ctx, input)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, output)
		mockSplitSvc.AssertExpectations(t)
	})

	t.Run("Create error", func(t *testing.T) {
		// Arrange
		input := &model.HashCrackTaskInput{
			MaxLength: 5,
			Hash:      hex.EncodeToString(md5.New().Sum([]byte("hash"))),
		}
		expectedErr := errors.New("create failed")

		mockSplitSvc.On("Split", ctx, input.MaxLength, mock.Anything).Return(10, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedErr).Once()

		// Act
		output, err := service.CreateTask(ctx, input)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, output)
		mockRepo.AssertExpectations(t)
		mockSplitSvc.AssertExpectations(t)
	})
}

func Test_GetTaskStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
		require.Equal(t, model.HashCrackStatusReady, output.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Task not found", func(t *testing.T) {
		// Arrange
		taskID := "123"
		mockRepo.On("Get", mock.Anything, taskID).Return(nil, repository.ErrCrackTaskNotFound).Once()

		// Act
		output, err := service.GetTaskStatus(ctx, taskID)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, repository.ErrCrackTaskNotFound)
		require.Nil(t, output)
		mockRepo.AssertExpectations(t)
	})
}

func Test_SaveResultTask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		input := &model.HashCrackTaskWebhookInput{
			RequestID:  "123",
			PartNumber: 0,
			Answer: struct {
				Words []string `xml:"words"`
			}{
				Words: []string{"word1", "word2"},
			},
		}

		task := &entity.HashCrackTask{
			ID:        input.RequestID,
			PartCount: 0,
		}

		mockRepo.On("Get", mock.Anything, input.RequestID).Return(task, nil).Once()
		mockRepo.On("Update", mock.Anything, input.RequestID, mock.Anything).Return(nil).Once()

		// Act
		err := service.SaveResultTask(ctx, input)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Task not found", func(t *testing.T) {
		// Arrange
		input := &model.HashCrackTaskWebhookInput{
			RequestID: "123",
			Answer: struct {
				Words []string `xml:"words"`
			}{
				Words: []string{"word1", "word2"},
			},
		}

		mockRepo.On("Get", mock.Anything, input.RequestID).Return(nil, repository.ErrCrackTaskNotFound).Once()

		// Act
		err := service.SaveResultTask(ctx, input)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, repository.ErrCrackTaskNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update error", func(t *testing.T) {
		// Arrange
		input := &model.HashCrackTaskWebhookInput{
			RequestID: "123",
			Answer: struct {
				Words []string `xml:"words"`
			}{
				Words: []string{"word1", "word2"},
			},
		}

		task := &entity.HashCrackTask{
			ID:        input.RequestID,
			PartCount: 1,
		}

		expectedError := errors.New("update failed")

		mockRepo.On("Get", mock.Anything, input.RequestID).Return(task, nil).Once()
		mockRepo.On("Update", mock.Anything, input.RequestID, mock.Anything).Return(expectedError).Once()

		// Act
		err := service.SaveResultTask(ctx, input)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, expectedError)
		mockRepo.AssertExpectations(t)
	})
}

func TestFinishTasks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		tasks := []*entity.HashCrackTask{
			{
				ID:     "123",
				Status: "PENDING",
			},
		}

		mockRepo.On("GetAllFinished", mock.Anything).Return(tasks, nil).Once()
		mockRepo.On("Update", mock.Anything, "123", mock.Anything).Return(nil).Once()

		// Act
		err := service.FinishTasks(ctx)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("No tasks", func(t *testing.T) {
		// Arrange
		mockRepo.On("GetAllFinished", mock.Anything).Return([]*entity.HashCrackTask{}, nil).Once()

		// Act
		err := service.FinishTasks(ctx)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetAllFinished error", func(t *testing.T) {
		// Arrange
		expectedError := errors.New("get all finished failed")
		mockRepo.On("GetAllFinished", mock.Anything).Return(nil, expectedError).Once()

		// Act
		err := service.FinishTasks(ctx)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, expectedError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update error", func(t *testing.T) {
		// Arrange
		tasks := []*entity.HashCrackTask{
			{
				ID:     "123",
				Status: "PENDING",
			},
		}

		expectedError := errors.New("update failed")

		mockRepo.On("GetAllFinished", mock.Anything).Return(tasks, nil).Once()
		mockRepo.On("Update", mock.Anything, "123", mock.Anything).Return(expectedError).Once()

		// Act
		err := service.FinishTasks(ctx)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, expectedError)
		mockRepo.AssertExpectations(t)
	})
}

func TestFinishTask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		taskID := "123"
		task := &entity.HashCrackTask{
			ID:     taskID,
			Status: "PENDING",
		}

		mockRepo.On("Get", mock.Anything, taskID).Return(task, nil).Once()
		mockRepo.On("Update", mock.Anything, taskID, mock.Anything).Return(nil).Once()

		// Act
		err := service.FinishTask(ctx, taskID)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Task not found", func(t *testing.T) {
		// Arrange
		taskID := "123"
		mockRepo.On("Get", mock.Anything, taskID).Return(nil, repository.ErrCrackTaskNotFound).Once()

		// Act
		err := service.FinishTask(ctx, taskID)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, repository.ErrCrackTaskNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update error", func(t *testing.T) {
		// Arrange
		taskID := "123"
		task := &entity.HashCrackTask{
			ID:     taskID,
			Status: "PENDING",
		}

		expectedError := errors.New("update failed")

		mockRepo.On("Get", mock.Anything, taskID).Return(task, nil).Once()
		mockRepo.On("Update", mock.Anything, taskID, mock.Anything).Return(expectedError).Once()

		// Act
		err := service.FinishTask(ctx, taskID)

		// Assert
		require.Error(t, err)
		require.ErrorIs(t, err, expectedError)
		mockRepo.AssertExpectations(t)
	})
}
