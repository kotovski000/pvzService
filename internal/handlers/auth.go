package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"pvzService/internal/models"
	"pvzService/internal/processors"
)

type Handler struct {
	Processor *processors.Processor
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		Processor: processors.NewProcessor(db),
	}
}

// --- Handlers ---

func (h *Handler) DummyLoginHandler(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Role string `json:"role"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body format"})
		}

		token, err := h.Processor.DummyLogin(body.Role, secret)

		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(models.Token{Token: token})
	}
}

func (h *Handler) RegisterHandler(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body format"})
		}
		user, err := h.Processor.Register(body.Email, body.Password, body.Role, secret)

		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(models.Token{Token: user})
	}
}

func (h *Handler) LoginHandler(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body format"})
		}

		token, err := h.Processor.Login(body.Email, body.Password, secret)

		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(models.Token{Token: token})
	}
}
