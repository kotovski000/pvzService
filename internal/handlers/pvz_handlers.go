package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pvzService/internal/models"
	"pvzService/internal/processors"
)

type PVZHandlers struct {
	pvzProcessor *processors.PVZProcessor
}

func NewPVZHandlers(pvzProcessor *processors.PVZProcessor) *PVZHandlers {
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

		return c.Status(fiber.StatusCreated).JSON(pvz)
	}
}

func (h *PVZHandlers) GetPVZListHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		startDate := c.Query("startDate")
		endDate := c.Query("endDate")
		pageStr := c.Query("page", "1")
		limitStr := c.Query("limit", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid page number"})
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid limit"})
		}

		result, err := h.pvzProcessor.ListPVZsWithRelations(startDate, endDate, page, limit)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: err.Error()})
		}

		// Преобразуем в формат, аналогичный оригинальному обработчику
		var response []map[string]interface{}
		for _, pvzWithRelations := range result {
			item := map[string]interface{}{
				"pvz": pvzWithRelations.PVZ,
			}

			var receptionsList []map[string]interface{}
			for _, reception := range pvzWithRelations.Receptions {
				receptionsList = append(receptionsList, map[string]interface{}{
					"reception": reception.Reception,
					"products":  reception.Products,
				})
			}

			item["receptions"] = receptionsList
			response = append(response, item)
		}

		return c.JSON(response)
	}
}
