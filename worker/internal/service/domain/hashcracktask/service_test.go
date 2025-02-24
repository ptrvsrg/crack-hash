package hashcracktask_test

import (
	"context"
	"crypto/md5"
	"errors"
	"github.com/ptrvsrg/crack-hash/worker/internal/logging"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain/hashcracktask"
	mock2 "github.com/ptrvsrg/crack-hash/worker/internal/service/infrastructure/mock"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ptrvsrg/crack-hash/worker/config"
	workermodel "github.com/ptrvsrg/crack-hash/worker/pkg/model"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"
)

var (
	mockBruteForce *mock2.HashBruteForceMock
	cfg            config.ManagerConfig
	svc            domain.HashCrackTask

	ctx = context.Background()
)

func init() {
	logging.Setup(config.EnvDev)
}

func TestMain(m *testing.M) {
	router := gin.New()
	router.POST("/internal/api/manager/hash/crack/webhook", func(c *gin.Context) {
		log.Debug().Msg("handle hash crack webhook")
		c.JSON(http.StatusOK, nil)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	address := strings.Replace(server.URL, "http://", "", 1)

	mockBruteForce = new(mock2.HashBruteForceMock)
	cfg = config.ManagerConfig{Address: address}
	svc = hashcracktask.NewService(
		cfg,
		resty.New(),
		mockBruteForce,
	)

	m.Run()
}

func Test_ExecuteTask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
	})

	t.Run("BruteForceError", func(t *testing.T) {
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
	})
}
