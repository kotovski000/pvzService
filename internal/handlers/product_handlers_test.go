package handlers

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvzService/internal/models"
)

type MockProductProcessor struct {
	mock.Mock
}

func (m *MockProductProcessor) AddProduct(pvzID, productType string) (models.Product, error) {
	args := m.Called(pvzID, productType)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductProcessor) DeleteLastProduct(pvzID string) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func TestProductHandlers_AddProductHandler_Success(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockProductProcessor)
	handler := NewProductHandlers(mockProcessor)

	testUUID := uuid.NewString()
	expectedProduct := models.Product{
		ID:          uuid.NewString(),
		Type:        "электроника",
		ReceptionId: testUUID,
	}

	mockProcessor.On("AddProduct", testUUID, "электроника").Return(expectedProduct, nil)

	app.Post("/products", handler.AddProductHandler())

	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(
		`{"type":"электроника","pvzId":"`+testUUID+`"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}

func TestProductHandlers_AddProductHandler_InvalidUUID(t *testing.T) {
	app := fiber.New()
	handler := NewProductHandlers(nil)

	app.Post("/products", handler.AddProductHandler())

	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(
		`{"type":"электроника","pvzId":"invalid-uuid"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestProductHandlers_DeleteLastProductHandler_Success(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockProductProcessor)
	handler := NewProductHandlers(mockProcessor)

	testUUID := uuid.NewString()
	mockProcessor.On("DeleteLastProduct", testUUID).Return(nil)

	app.Post("/pvz/:pvzId/delete_last_product", handler.DeleteLastProductHandler())

	req := httptest.NewRequest("POST", "/pvz/"+testUUID+"/delete_last_product", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}
