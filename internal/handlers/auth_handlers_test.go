package handlers

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthProcessor struct {
	mock.Mock
}

func (m *MockAuthProcessor) Register(email, password, role string) (string, error) {
	args := m.Called(email, password, role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthProcessor) Login(email, password string) (string, string, error) {
	args := m.Called(email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthProcessor) DummyLogin(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthProcessor) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthProcessor) ComparePassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

func TestAuthHandlers_DummyLoginHandler_Success(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockAuthProcessor)
	handler := NewAuthHandlers(mockProcessor, "secret")

	mockProcessor.On("DummyLogin", "employee").Return("user123", nil)

	app.Post("/dummyLogin", handler.DummyLoginHandler())

	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBufferString(`{"role":"employee"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}

func TestAuthHandlers_DummyLoginHandler_InvalidRole(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockAuthProcessor)
	handler := NewAuthHandlers(mockProcessor, "secret")

	mockProcessor.On("DummyLogin", "invalid").Return("", errors.New("invalid role"))

	app.Post("/dummyLogin", handler.DummyLoginHandler())

	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBufferString(`{"role":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}

func TestAuthHandlers_RegisterHandler_Success(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockAuthProcessor)
	handler := NewAuthHandlers(mockProcessor, "secret")

	mockProcessor.On("Register", "test@example.com", "password", "employee").Return("user123", nil)

	app.Post("/register", handler.RegisterHandler())

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(`{"email":"test@example.com","password":"password","role":"employee"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}

func TestAuthHandlers_RegisterHandler_InvalidRole(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockAuthProcessor)
	handler := NewAuthHandlers(mockProcessor, "secret")

	mockProcessor.On("Register", "test@example.com", "password", "invalid").Return("", errors.New("invalid role"))

	app.Post("/register", handler.RegisterHandler())

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(`{"email":"test@example.com","password":"password","role":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}

func TestAuthHandlers_RegisterHandler_EmailExists(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockAuthProcessor)
	handler := NewAuthHandlers(mockProcessor, "secret")

	mockProcessor.On("Register", "exists@example.com", "password", "employee").Return("", errors.New("email already exists"))

	app.Post("/register", handler.RegisterHandler())

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(`{"email":"exists@example.com","password":"password","role":"employee"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}

func TestAuthHandlers_LoginHandler_Success(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockAuthProcessor)
	handler := NewAuthHandlers(mockProcessor, "secret")

	mockProcessor.On("Login", "test@example.com", "password").Return("user123", "employee", nil)

	app.Post("/login", handler.LoginHandler())

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email":"test@example.com","password":"password"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}

func TestAuthHandlers_LoginHandler_InvalidCredentials(t *testing.T) {
	app := fiber.New()
	mockProcessor := new(MockAuthProcessor)
	handler := NewAuthHandlers(mockProcessor, "secret")

	mockProcessor.On("Login", "test@example.com", "wrong").Return("", "", errors.New("invalid email or password"))

	app.Post("/login", handler.LoginHandler())

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email":"test@example.com","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	mockProcessor.AssertExpectations(t)
}

func TestGenerateToken(t *testing.T) {
	handler := NewAuthHandlers(nil, "secret")
	token, err := handler.GenerateToken("user123", "employee")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, "user123", claims["userId"])
	assert.Equal(t, "employee", claims["role"])
	assert.InDelta(t, time.Now().Add(24*time.Hour).Unix(), claims["exp"].(float64), 10)
}
