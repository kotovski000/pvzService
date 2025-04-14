package processors

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pvzService/internal/models"
)

type MockReceptionRepository struct {
	mock.Mock
}

func (m *MockReceptionRepository) CreateReception(pvzID string, idGenerator func() uuid.UUID) (string, error) {
	args := m.Called(pvzID, idGenerator)
	return args.String(0), args.Error(1)
}

func (m *MockReceptionRepository) GetReceptionByID(id string) (models.Reception, error) {
	args := m.Called(id)
	return args.Get(0).(models.Reception), args.Error(1)
}

func (m *MockReceptionRepository) GetOpenReception(pvzID string) (models.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

func (m *MockReceptionRepository) CloseReception(id string, closeTime time.Time) error {
	args := m.Called(id, closeTime)
	return args.Error(0)
}

func (m *MockReceptionRepository) HasOpenReception(pvzID string) (bool, error) {
	args := m.Called(pvzID)
	return args.Bool(0), args.Error(1)
}

func TestReceptionProcessor_CreateReception(t *testing.T) {
	mockRepo := new(MockReceptionRepository)
	processor := NewReceptionProcessor(mockRepo)

	t.Run("success", func(t *testing.T) {
		pvzID := uuid.New().String()
		receptionID := uuid.New().String()
		expectedReception := models.Reception{
			ID:       receptionID,
			PvzId:    pvzID,
			Status:   "in_progress",
			DateTime: time.Now(),
		}

		mockRepo.On("HasOpenReception", pvzID).Return(false, nil)
		mockRepo.On("CreateReception", pvzID, mock.AnythingOfType("func() uuid.UUID")).Return(receptionID, nil)
		mockRepo.On("GetReceptionByID", receptionID).Return(expectedReception, nil)

		result, err := processor.CreateReception(pvzID)
		assert.NoError(t, err)
		assert.Equal(t, expectedReception, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("has open reception", func(t *testing.T) {
		pvzID := uuid.New().String()
		mockRepo.On("HasOpenReception", pvzID).Return(true, nil)

		_, err := processor.CreateReception(pvzID)
		assert.EqualError(t, err, "open reception already exists for this PVZ")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on has open", func(t *testing.T) {
		pvzID := uuid.New().String()
		mockRepo.On("HasOpenReception", pvzID).Return(false, errors.New("db error"))

		_, err := processor.CreateReception(pvzID)
		assert.EqualError(t, err, "database error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on create", func(t *testing.T) {
		pvzID := uuid.New().String()
		mockRepo.On("HasOpenReception", pvzID).Return(false, nil)
		mockRepo.On("CreateReception", pvzID, mock.AnythingOfType("func() uuid.UUID")).Return("", errors.New("db error"))

		_, err := processor.CreateReception(pvzID)
		assert.EqualError(t, err, "failed to create reception")
		mockRepo.AssertExpectations(t)
	})
}

func TestReceptionProcessor_CloseLastReception(t *testing.T) {
	mockRepo := new(MockReceptionRepository)
	processor := NewReceptionProcessor(mockRepo)

	t.Run("success", func(t *testing.T) {
		pvzID := uuid.New().String()
		receptionID := uuid.New().String()
		now := time.Now()
		openReception := models.Reception{
			ID:       receptionID,
			PvzId:    pvzID,
			Status:   "in_progress",
			DateTime: time.Now(),
		}
		expectedReception := openReception
		expectedReception.Status = "close"
		expectedReception.ClosedAt = &now

		mockRepo.On("GetOpenReception", pvzID).Return(openReception, nil)
		mockRepo.On("CloseReception", receptionID, mock.AnythingOfType("time.Time")).Return(nil)

		result, err := processor.CloseLastReception(pvzID)
		assert.NoError(t, err)
		assert.Equal(t, expectedReception.Status, result.Status)
		assert.NotNil(t, result.ClosedAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no open reception", func(t *testing.T) {
		pvzID := uuid.New().String()
		mockRepo.On("GetOpenReception", pvzID).Return(models.Reception{}, sql.ErrNoRows)

		_, err := processor.CloseLastReception(pvzID)
		assert.EqualError(t, err, "no open reception found for this PVZ")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on get open", func(t *testing.T) {
		pvzID := uuid.New().String()
		mockRepo.On("GetOpenReception", pvzID).Return(models.Reception{}, errors.New("db error"))

		_, err := processor.CloseLastReception(pvzID)
		assert.EqualError(t, err, "database error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on close", func(t *testing.T) {
		pvzID := uuid.New().String()
		receptionID := uuid.New().String()
		openReception := models.Reception{
			ID:       receptionID,
			PvzId:    pvzID,
			Status:   "in_progress",
			DateTime: time.Now(),
		}

		mockRepo.On("GetOpenReception", pvzID).Return(openReception, nil)
		mockRepo.On("CloseReception", receptionID, mock.AnythingOfType("time.Time")).Return(errors.New("db error"))

		_, err := processor.CloseLastReception(pvzID)
		assert.EqualError(t, err, "failed to close reception")
		mockRepo.AssertExpectations(t)
	})
}
