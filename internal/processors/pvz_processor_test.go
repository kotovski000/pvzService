package processors

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvzService/internal/models"
	"pvzService/internal/repository"
)

type MockPVZRepo struct {
	mock.Mock
}

func (m *MockPVZRepo) CreatePVZ(city string, idGenerator func() uuid.UUID) (models.PVZ, error) {
	args := m.Called(city, idGenerator)
	return args.Get(0).(models.PVZ), args.Error(1)
}

func (m *MockPVZRepo) GetPVZByID(id string) (models.PVZ, error) {
	args := m.Called(id)
	return args.Get(0).(models.PVZ), args.Error(1)
}

func (m *MockPVZRepo) ListPVZsWithRelations(startDate, endDate time.Time, limit, offset int) ([]repository.PVZResponse, error) {
	args := m.Called(startDate, endDate, limit, offset)
	return args.Get(0).([]repository.PVZResponse), args.Error(1)
}

func TestPVZProcessor_CreatePVZ(t *testing.T) {
	mockRepo := new(MockPVZRepo)
	processor := NewPVZProcessor(mockRepo)

	t.Run("success", func(t *testing.T) {
		expectedPVZ := models.PVZ{
			ID:   uuid.NewString(),
			City: "Москва",
		}

		mockRepo.On("CreatePVZ", "Москва", mock.AnythingOfType("func() uuid.UUID")).
			Return(expectedPVZ, nil)

		pvz, err := processor.CreatePVZ("Москва")

		assert.NoError(t, err)
		assert.Equal(t, "Москва", pvz.City)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid city", func(t *testing.T) {
		_, err := processor.CreatePVZ("Нью-Йорк")
		assert.Error(t, err)
		assert.Equal(t, "invalid city", err.Error())
	})
}

func TestPVZProcessor_GetPVZByID(t *testing.T) {
	mockRepo := new(MockPVZRepo)
	processor := NewPVZProcessor(mockRepo)

	t.Run("success", func(t *testing.T) {
		expectedPVZ := models.PVZ{
			ID:   "test-id",
			City: "Москва",
		}

		mockRepo.On("GetPVZByID", "test-id").Return(expectedPVZ, nil)

		pvz, err := processor.GetPVZByID("test-id")

		assert.NoError(t, err)
		assert.Equal(t, "Москва", pvz.City)
		mockRepo.AssertExpectations(t)
	})
}

func TestPVZProcessor_ListPVZsWithRelations(t *testing.T) {
	mockRepo := new(MockPVZRepo)
	processor := NewPVZProcessor(mockRepo)

	t.Run("success", func(t *testing.T) {
		expected := []repository.PVZResponse{
			{
				PVZ: models.PVZ{
					ID:   "pvz1",
					City: "Москва",
				},
			},
		}

		mockRepo.On("ListPVZsWithRelations", time.Time{}, time.Time{}, 10, 0).
			Return(expected, nil)

		result, err := processor.ListPVZsWithRelations("", "", 1, 10)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid date format", func(t *testing.T) {
		_, err := processor.ListPVZsWithRelations("invalid", "", 1, 10)
		assert.Error(t, err)
	})

	t.Run("invalid pagination", func(t *testing.T) {
		_, err := processor.ListPVZsWithRelations("", "", 0, 10)
		assert.Error(t, err)
	})
}
