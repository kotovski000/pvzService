package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"pvzService/internal/middleware"

	"pvzService/internal/models"
	"pvzService/internal/processors"
)

type ReceptionHandlers struct {
	receptionProcessor *processors.ReceptionProcessor
}

func NewReceptionHandlers(receptionProcessor *processors.ReceptionProcessor) *ReceptionHandlers {
	return &ReceptionHandlers{receptionProcessor: receptionProcessor}
}

func (h *ReceptionHandlers) CreateReceptionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			PvzId string `json:"pvzId"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request"})
		}

		if _, err := uuid.Parse(body.PvzId); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid pvzId format"})
		}

		reception, err := h.receptionProcessor.CreateReception(body.PvzId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: err.Error()})
		}

		middleware.OrderAcceptancesCreated.Inc()
		return c.Status(fiber.StatusCreated).JSON(reception)
	}
}

func (h *ReceptionHandlers) CloseLastReceptionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		pvzId := c.Params("pvzId")

		if _, err := uuid.Parse(pvzId); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid pvzId format"})
		}

		reception, err := h.receptionProcessor.CloseLastReception(pvzId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: err.Error()})
		}

		return c.JSON(reception)
	}
}
