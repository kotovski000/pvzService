package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"pvzService/internal/models"
	"time"
)

// --- Authentication ---

func GenerateToken(userID, role string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
		"iat":    time.Now().Unix(),
		"nbf":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func comparePassword(hashedPassword string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err
}

// --- Handlers ---

func DummyLoginHandler(db *sql.DB, secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Role string `json:"role"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body"})
		}

		if body.Role != "employee" && body.Role != "moderator" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid role"})
		}

		// Creating a dummy user
		var userID string
		err := db.QueryRow("SELECT id FROM users WHERE role = $1 LIMIT 1", body.Role).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Create a dummy user if none exists
				userID = uuid.New().String()
				_, err = db.Exec("INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)", userID, "dummy@example.com", "password", body.Role)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to create dummy user"})
				}
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Database error"})
			}
		}

		token, err := GenerateToken(userID, body.Role, secret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to generate token"})
		}

		return c.JSON(models.Token{Token: token})
	}
}

func RegisterHandler(db *sql.DB, secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body"})
		}

		if body.Role != "employee" && body.Role != "moderator" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid role"})
		}

		hashedPassword, err := hashPassword(body.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to hash password"})
		}

		userID := uuid.New().String()

		_, err = db.Exec("INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)",
			userID, body.Email, hashedPassword, body.Role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: fmt.Sprintf("Failed to create user: %v", err)})
		}

		token, err := GenerateToken(userID, body.Role, secret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to generate token"})
		}

		return c.Status(fiber.StatusCreated).JSON(models.Token{Token: token})
	}
}

func LoginHandler(db *sql.DB, secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body"})
		}

		var userID, hashedPassword, role string
		err := db.QueryRow("SELECT id, password, role FROM users WHERE email = $1", body.Email).Scan(&userID, &hashedPassword, &role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusUnauthorized).JSON(models.Error{Message: "Invalid credentials"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Database error"})
		}

		if err := comparePassword(hashedPassword, body.Password); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.Error{Message: "Invalid credentials"})
		}

		token, err := GenerateToken(userID, role, secret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to generate token"})
		}

		return c.JSON(models.Token{Token: token})
	}
}
