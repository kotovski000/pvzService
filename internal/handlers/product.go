package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"pvzService/internal/models"
)

func (h *Handler) AddProductHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Type  string `json:"type"`
			PvzId string `json:"pvzId"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body format"})
		}

		product, err := h.Processor.AddProduct(body.Type, body.PvzId)

		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(product)
	}
}

func (h *Handler) DeleteLastProductHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		pvzId := c.Params("pvzId")

		err := h.Processor.DeleteLastProduct(pvzId)

		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: err.Error()})
		}

		return c.SendStatus(fiber.StatusOK)
	}
}
