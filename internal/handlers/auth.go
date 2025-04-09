package handlers

import (
	"database/sql"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"pvzService/internal/models"
	"time"
)

func GenerateToken(userID, role string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
		"iat":    time.Now().Unix(),
		"nbf":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func comparePassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func DummyLoginHandler(db *sql.DB, secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Role string `json:"role"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body format"})
		}

		if body.Role != "employee" && body.Role != "moderator" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "Invalid role. Allowed values: employee, moderator",
			})
		}

		var userID string
		err := db.QueryRow("SELECT id FROM users WHERE role = $1 LIMIT 1", body.Role).Scan(&userID)

		if errors.Is(err, sql.ErrNoRows) {
			hashedPassword, err := hashPassword("password")
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
					Message: "Failed to create dummy user",
				})
			}

			userID = uuid.New().String()
			_, err = db.Exec(
				"INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)",
				userID, "dummy@example.com", hashedPassword, body.Role,
			)

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
					Message: "Failed to create dummy user: " + err.Error(),
				})
			}
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Database error: " + err.Error(),
			})
		}

		token, err := GenerateToken(userID, body.Role, secret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Failed to generate token: " + err.Error(),
			})
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
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "Invalid request body format",
			})
		}

		if body.Role != "employee" && body.Role != "moderator" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "Invalid role. Allowed values: employee, moderator",
			})
		}

		hashedPassword, err := hashPassword(body.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Failed to process password",
			})
		}

		userID := uuid.New().String()
		_, err = db.Exec(
			"INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)",
			userID, body.Email, hashedPassword, body.Role,
		)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{
					Message: "Email already exists",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Failed to create user: " + err.Error(),
			})
		}

		token, err := GenerateToken(userID, body.Role, secret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Failed to generate token: " + err.Error(),
			})
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
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "Invalid request body format",
			})
		}

		var userID, hashedPassword, role string
		err := db.QueryRow(
			"SELECT id, password, role FROM users WHERE email = $1",
			body.Email,
		).Scan(&userID, &hashedPassword, &role)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusUnauthorized).JSON(models.Error{
					Message: "Invalid email or password",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Database error: " + err.Error(),
			})
		}

		if err := comparePassword(hashedPassword, body.Password); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.Error{
				Message: "Invalid email or password",
			})
		}

		token, err := GenerateToken(userID, role, secret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: "Failed to generate token: " + err.Error(),
			})
		}

		return c.JSON(models.Token{Token: token})
	}
}
