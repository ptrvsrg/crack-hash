package hashcracktask_test

import (
	"context"
	"crypto/md5"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jobqueue "github.com/dirkaholic/kyoo"
	"github.com/rs/zerolog/log"

	"github.com/ptrvsrg/crack-hash/commonlib/http/client"
	"github.com/ptrvsrg/crack-hash/commonlib/logging"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain/hashcracktask"
	mock2 "github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"

	"github.com/ptrvsrg/crack-hash/worker/config"
	workermodel "github.com/ptrvsrg/crack-hash/worker/pkg/model"
)

var (
	mockBruteForce *mock2.HashBruteForceMock
	cfg            config.ManagerConfig
	svc            domain.HashCrackTask

	ctx = context.Background()
)

func init() {
	logging.Setup(true)
}

func TestMain(m *testing.M) {
	// HTTP server
	router := gin.New()
	router.POST(
		"/internal/api/manager/hash/crack/webhook", func(c *gin.Context) {
			log.Debug().Msg("handle hash crack webhook")
			c.JSON(http.StatusOK, nil)
		},
	)

	server := httptest.NewServer(router)
	defer server.Close()

	// HTTP client
	httpClient, err := client.New(
		client.WithRetries(3, 5*time.Second, 10*time.Second),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create HTTP client")
	}
	defer func(httpClient *resty.Client) {
		err := httpClient.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close HTTP client")
		}
	}(httpClient)

	// Job queue
	jobQueue := jobqueue.NewJobQueue(1000)
	jobQueue.Start()
	defer jobQueue.Stop()

	// Service
	mockBruteForce = new(mock2.HashBruteForceMock)
	cfg = config.ManagerConfig{Address: server.URL}
	svc = hashcracktask.NewService(
		cfg,
		httpClient,
		jobQueue,
		mockBruteForce,
	)

	m.Run()
}

func Test_ExecuteTask(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			// Arrange
			hash := md5.Sum([]byte("abc"))
			input := &workermodel.HashCrackTaskInput{
				RequestID:  "123",
				Hash:       string(hash[:]),
				MaxLength:  5,
				PartNumber: 0,
				Alphabet: struct {
					Symbols []string `xml:"Symbols"`
				}{
					Symbols: []string{"a", "b", "c"},
				},
			}

			mockBruteForce.On("BruteForceMD5", input.Hash, input.Alphabet.Symbols, input.MaxLength, input.PartNumber).
				Return([]string{"abc"}, nil).Once()

			// Act
			err := svc.ExecuteTask(context.Background(), input)
			time.Sleep(time.Second)

			// Assert
			require.NoError(t, err)
			mockBruteForce.AssertExpectations(t)
		},
	)

	t.Run(
		"BruteForceError", func(t *testing.T) {
			// Arrange
			hash := md5.Sum([]byte("abc"))
			input := &workermodel.HashCrackTaskInput{
				RequestID:  "123",
				Hash:       string(hash[:]),
				MaxLength:  5,
				PartNumber: 0,
				Alphabet: struct {
					Symbols []string `xml:"Symbols"`
				}{
					Symbols: []string{"a", "b", "c"},
				},
			}
			expectedError := errors.New("brute force failed")

			mockBruteForce.On("BruteForceMD5", input.Hash, input.Alphabet.Symbols, input.MaxLength, input.PartNumber).
				Return(nil, expectedError).Once()

			// Act
			err := svc.ExecuteTask(context.Background(), input)
			time.Sleep(time.Second)

			// Assert
			require.NoError(t, err)
			mockBruteForce.AssertExpectations(t)
		},
	)
}
