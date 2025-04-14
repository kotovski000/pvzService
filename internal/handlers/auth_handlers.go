package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"time"

	"pvzService/internal/models"
	"pvzService/internal/processors"
)

type AuthHandlers struct {
	authProcessor processors.AuthProcessor
	secret        string
}

func NewAuthHandlers(authProcessor processors.AuthProcessor, secret string) *AuthHandlers {
	return &AuthHandlers{
		authProcessor: authProcessor,
		secret:        secret,
	}
}

func (h *AuthHandlers) GenerateToken(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * 1).Unix(),
		"iat":    time.Now().Unix(),
		"nbf":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.secret))
}

func (h *AuthHandlers) DummyLoginHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Role string `json:"role"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body format"})
		}

		userID, err := h.authProcessor.DummyLogin(body.Role)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: err.Error(),
			})
		}

		token, err := h.GenerateToken(userID, body.Role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Failed to generate token: " + err.Error(),
			})
		}

		return c.JSON(models.Token{Token: token})
	}
}

func (h *AuthHandlers) RegisterHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "Invalid request body format",
			})
		}

		userID, err := h.authProcessor.Register(body.Email, body.Password, body.Role)
		if err != nil {
			status := fiber.StatusInternalServerError
			if err.Error() == "invalid role" || err.Error() == "email already exists" {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(models.Error{
				Message: err.Error(),
			})
		}

		token, err := h.GenerateToken(userID, body.Role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Failed to generate token: " + err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(models.Token{Token: token})
	}
}

func (h *AuthHandlers) LoginHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "Invalid request body format",
			})
		}

		userID, role, err := h.authProcessor.Login(body.Email, body.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.Error{
				Message: err.Error(),
			})
		}

		token, err := h.GenerateToken(userID, role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Failed to generate token: " + err.Error(),
			})
		}

		return c.JSON(models.Token{Token: token})
	}
}
