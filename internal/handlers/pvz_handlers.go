package handlers

import (
	"pvzService/internal/prometheus"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"pvzService/internal/models"
	"pvzService/internal/processors"
)

type PVZHandlers struct {
	pvzProcessor processors.PVZProcessor
}

func NewPVZHandlers(pvzProcessor processors.PVZProcessor) *PVZHandlers {
	return &PVZHandlers{pvzProcessor: pvzProcessor}
}

func (h *PVZHandlers) CreatePVZHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body models.PVZ
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request"})
		}

		pvz, err := h.pvzProcessor.CreatePVZ(body.City)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: err.Error()})
		}

		prometheus.PickupPointsCreated.Inc()

		return c.Status(fiber.StatusCreated).JSON(pvz)

	}
}

func (h *PVZHandlers) GetPVZListHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		pageStr := c.Query("page")
		if pageStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "page parameter is required",
			})
		}

		limitStr := c.Query("limit")
		if limitStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "limit parameter is required",
			})
		}

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "page must be a positive integer",
			})
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 30 {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Message: "limit must be between 1 and 30",
			})
		}

		startDate := c.Query("startDate")
		if startDate != "" {
			if _, err := time.Parse(time.RFC3339, startDate); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{
					Message: "invalid startDate format, must be RFC3339",
				})
			}
		}

		endDate := c.Query("endDate")
		if endDate != "" {
			if _, err := time.Parse(time.RFC3339, endDate); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{
					Message: "invalid endDate format, must be RFC3339",
				})
			}
		}

		result, err := h.pvzProcessor.ListPVZsWithRelations(startDate, endDate, page, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
				Message: err.Error(),
			})
		}

		return c.JSON(result)
	}
}
