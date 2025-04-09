package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"pvzService/internal/models"
)

func (h *Handler) CreateReceptionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			PvzId string `json:"pvzId"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body format"})
		}

		claims, ok := c.Locals("claims").(map[string]interface{})
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to get claims from context"})
		}

		role, ok := claims["role"].(string)
		if !ok || role != "employee" {
			return c.Status(fiber.StatusForbidden).JSON(models.Error{Message: "Insufficient role"})
		}

		reception, err := h.Processor.CreateReception(body.PvzId)

		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(reception)
	}
}

func (h *Handler) CloseLastReceptionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		pvzId := c.Params("pvzId")

		claims, ok := c.Locals("claims").(map[string]interface{})
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to get claims from context"})
		}

		role, ok := claims["role"].(string)
		if !ok || role != "employee" {
			return c.Status(fiber.StatusForbidden).JSON(models.Error{Message: "Insufficient role"})
		}

		reception, err := h.Processor.CloseLastReception(pvzId)
		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(reception)
	}
}
