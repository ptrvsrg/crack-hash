package chunkbased_test

import (
	"context"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/logging"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/tasksplit/chunkbased"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	svc infrastructure.TaskSplit

	ctx = context.Background()
)

func init() {
	logging.Setup(config.EnvDev)
}

func TestMain(m *testing.M) {
	svc = chunkbased.NewService()

	m.Run()
}

func TestSplit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		wordMaxLength := 3
		alphabetLength := 5

		// Act
		numSubtasks, err := svc.Split(ctx, wordMaxLength, alphabetLength)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 1, numSubtasks) // Ожидаемое количество подзадач
	})

	t.Run("Large word count", func(t *testing.T) {
		// Arrange
		wordMaxLength := 10
		alphabetLength := 26

		// Act
		numSubtasks, err := svc.Split(ctx, wordMaxLength, alphabetLength)

		// Assert
		require.NoError(t, err)
		assert.Greater(t, numSubtasks, 1) // Количество подзадач должно быть больше 1
	})

	t.Run("Invalid input", func(t *testing.T) {
		t.Run("Invalid wordMaxLength", func(t *testing.T) {
			// Arrange
			wordMaxLength := -1
			alphabetLength := 5

			// Act
			numSubtasks, err := svc.Split(ctx, wordMaxLength, alphabetLength)

			// Assert
			require.Error(t, err)
			assert.ErrorIs(t, err, infrastructure.ErrInvalidWordMaxLength)
			assert.Equal(t, 0, numSubtasks)
		})

		t.Run("Invalid alphabetLength", func(t *testing.T) {
			// Arrange
			wordMaxLength := 3
			alphabetLength := -1

			// Act
			numSubtasks, err := svc.Split(ctx, wordMaxLength, alphabetLength)

			// Assert
			require.Error(t, err)
			assert.ErrorIs(t, err, infrastructure.ErrInvalidAlphabetLength)
			assert.Equal(t, 0, numSubtasks)
		})
	})
}
