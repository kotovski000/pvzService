package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"pvzService/internal/models"
)

func (h *Handler) CreatePVZHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			City string `json:"city"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body format"})
		}
		pvz, err := h.Processor.CreatePVZ(body.City)

		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(pvz)
	}
}

func (h *Handler) GetPVZListHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		startDate := c.Query("startDate")
		endDate := c.Query("endDate")
		page := c.Query("page")
		limit := c.Query("limit")

		pvzs, err := h.Processor.GetPVZList(startDate, endDate, page, limit)

		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(pvzs)
	}
}
