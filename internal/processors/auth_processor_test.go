package processors

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateUser(email, hashedPassword, role string) (string, error) {
	args := m.Called(email, hashedPassword, role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthRepository) FindUserByEmail(email string) (string, string, string, error) {
	args := m.Called(email)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockAuthRepository) FindUserByRole(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func TestAuthProcessor_Register_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	mockRepo.On("CreateUser", "test@example.com", mock.Anything, "employee").Return("user123", nil)

	userID, err := processor.Register("test@example.com", "password", "employee")
	assert.NoError(t, err)
	assert.Equal(t, "user123", userID)
	mockRepo.AssertExpectations(t)
}

func TestAuthProcessor_Register_InvalidRole(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	_, err := processor.Register("test@example.com", "password", "invalid")
	assert.Error(t, err)
	assert.Equal(t, "invalid role", err.Error())
}

func TestAuthProcessor_Register_EmailExists(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	mockRepo.On("CreateUser", "exists@example.com", mock.Anything, "employee").Return("", errors.New("email already exists"))

	_, err := processor.Register("exists@example.com", "password", "employee")
	assert.Error(t, err)
	assert.Equal(t, "email already exists", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestAuthProcessor_Login_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	hashedPassword, _ := processor.HashPassword("password")
	mockRepo.On("FindUserByEmail", "test@example.com").Return("user123", hashedPassword, "employee", nil)

	userID, role, err := processor.Login("test@example.com", "password")
	assert.NoError(t, err)
	assert.Equal(t, "user123", userID)
	assert.Equal(t, "employee", role)
	mockRepo.AssertExpectations(t)
}

func TestAuthProcessor_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	hashedPassword, _ := processor.HashPassword("password")
	mockRepo.On("FindUserByEmail", "test@example.com").Return("user123", hashedPassword, "employee", nil)

	_, _, err := processor.Login("test@example.com", "wrong")
	assert.Error(t, err)
	assert.Equal(t, "invalid email or password", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestAuthProcessor_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	mockRepo.On("FindUserByEmail", "nonexistent@example.com").Return("", "", "", sql.ErrNoRows)

	_, _, err := processor.Login("nonexistent@example.com", "password")
	assert.Error(t, err)
	assert.Equal(t, "invalid email or password", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestAuthProcessor_DummyLogin_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	mockRepo.On("FindUserByRole", "employee").Return("user123", nil)

	userID, err := processor.DummyLogin("employee")
	assert.NoError(t, err)
	assert.Equal(t, "user123", userID)
	mockRepo.AssertExpectations(t)
}

func TestAuthProcessor_DummyLogin_CreateNewUser(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	mockRepo.On("FindUserByRole", "employee").Return("", sql.ErrNoRows)
	mockRepo.On("CreateUser", "dummy@example.com", mock.Anything, "employee").Return("newuser123", nil)

	userID, err := processor.DummyLogin("employee")
	assert.NoError(t, err)
	assert.Equal(t, "newuser123", userID)
	mockRepo.AssertExpectations(t)
}

func TestAuthProcessor_DummyLogin_InvalidRole(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	processor := NewAuthProcessor(mockRepo)

	_, err := processor.DummyLogin("invalid")
	assert.Error(t, err)
	assert.Equal(t, "invalid role", err.Error())
}

func TestHashAndComparePassword(t *testing.T) {
	processor := NewAuthProcessor(nil)
	password := "testpassword123"

	hashed, err := processor.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)

	err = processor.ComparePassword(hashed, password)
	assert.NoError(t, err)

	err = processor.ComparePassword(hashed, "wrongpassword")
	assert.Error(t, err)
}
