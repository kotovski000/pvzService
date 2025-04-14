package handlers

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pvzService/internal/models"
)

type MockReceptionProcessor struct {
	mock.Mock
}

func (m *MockReceptionProcessor) CreateReception(pvzID string) (models.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

func (m *MockReceptionProcessor) CloseLastReception(pvzID string) (models.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

func TestReceptionHandlers_CreateReceptionHandler(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockReceptionProcessor)
	handler := NewReceptionHandlers(mockProcessor)

	t.Run("success", func(t *testing.T) {
		pvzID := uuid.New().String()
		expectedReception := models.Reception{
			ID:       uuid.New().String(),
			PvzId:    pvzID,
			Status:   "in_progress",
			DateTime: time.Now(),
		}

		mockProcessor.On("CreateReception", pvzID).Return(expectedReception, nil)

		app.Post("/receptions", handler.CreateReceptionHandler())

		reqBody := `{"pvzId":"` + pvzID + `"}`
		req := httptest.NewRequest("POST", "/receptions", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
		mockProcessor.AssertExpectations(t)
	})

	t.Run("invalid pvzId format", func(t *testing.T) {
		app.Post("/receptions", handler.CreateReceptionHandler())

		reqBody := `{"pvzId":"invalid-uuid"}`
		req := httptest.NewRequest("POST", "/receptions", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid request body", func(t *testing.T) {
		app.Post("/receptions", handler.CreateReceptionHandler())

		req := httptest.NewRequest("POST", "/receptions", bytes.NewBufferString("invalid"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestReceptionHandlers_CloseLastReceptionHandler(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockReceptionProcessor)
	handler := NewReceptionHandlers(mockProcessor)

	t.Run("success", func(t *testing.T) {
		pvzID := uuid.New().String()
		expectedReception := models.Reception{
			ID:       uuid.New().String(),
			PvzId:    pvzID,
			Status:   "close",
			DateTime: time.Now(),
			ClosedAt: &time.Time{},
		}

		mockProcessor.On("CloseLastReception", pvzID).Return(expectedReception, nil)

		app.Put("/receptions/:pvzId/close", handler.CloseLastReceptionHandler())

		req := httptest.NewRequest("PUT", "/receptions/"+pvzID+"/close", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockProcessor.AssertExpectations(t)
	})

	t.Run("invalid pvzId format", func(t *testing.T) {
		app.Put("/receptions/:pvzId/close", handler.CloseLastReceptionHandler())

		req := httptest.NewRequest("PUT", "/receptions/invalid-uuid/close", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}
