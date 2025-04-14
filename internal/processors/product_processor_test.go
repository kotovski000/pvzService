package processors

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvzService/internal/models"
)

type MockProductRepo struct {
	mock.Mock
}

func (m *MockProductRepo) AddProduct(receptionID, productType string, idGenerator func() uuid.UUID) (string, error) {
	args := m.Called(receptionID, productType, idGenerator)
	return args.String(0), args.Error(1)
}

func (m *MockProductRepo) GetProductByID(id string) (models.Product, error) {
	args := m.Called(id)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductRepo) GetLastProduct(receptionID string) (models.Product, error) {
	args := m.Called(receptionID)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductRepo) DeleteProduct(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockReceptionRepo struct {
	mock.Mock
}

func (m *MockReceptionRepo) GetOpenReception(pvzID string) (models.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

func TestProductProcessor_AddProduct_Success(t *testing.T) {
	mockProductRepo := new(MockProductRepo)
	mockReceptionRepo := new(MockReceptionRepo)
	processor := NewProductProcessor(mockProductRepo, mockReceptionRepo)

	pvzID := uuid.NewString()
	receptionID := uuid.NewString()
	productID := uuid.NewString()

	mockReceptionRepo.On("GetOpenReception", pvzID).Return(
		models.Reception{ID: receptionID}, nil)

	mockProductRepo.On("AddProduct", receptionID, "электроника", mock.AnythingOfType("func() uuid.UUID")).
		Return(productID, nil)

	mockProductRepo.On("GetProductByID", productID).Return(
		models.Product{ID: productID, Type: "электроника"}, nil)

	product, err := processor.AddProduct(pvzID, "электроника")
	assert.NoError(t, err)
	assert.Equal(t, "электроника", product.Type)
	mockProductRepo.AssertExpectations(t)
	mockReceptionRepo.AssertExpectations(t)
}

func TestProductProcessor_DeleteLastProduct_Success(t *testing.T) {
	mockProductRepo := new(MockProductRepo)
	mockReceptionRepo := new(MockReceptionRepo)
	processor := NewProductProcessor(mockProductRepo, mockReceptionRepo)

	pvzID := uuid.NewString()
	receptionID := uuid.NewString()
	productID := uuid.NewString()

	mockReceptionRepo.On("GetOpenReception", pvzID).Return(
		models.Reception{ID: receptionID}, nil)
	mockProductRepo.On("GetLastProduct", receptionID).Return(
		models.Product{ID: productID}, nil)
	mockProductRepo.On("DeleteProduct", productID).Return(nil)

	err := processor.DeleteLastProduct(pvzID)
	assert.NoError(t, err)
	mockProductRepo.AssertExpectations(t)
	mockReceptionRepo.AssertExpectations(t)
}
