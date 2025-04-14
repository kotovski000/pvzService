package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"pvzService/internal/prometheus"

	"pvzService/internal/models"
)

type ProductProcessor interface {
	AddProduct(pvzID, productType string) (models.Product, error)
	DeleteLastProduct(pvzID string) error
}

type ProductHandlers struct {
	productProcessor ProductProcessor
}

func NewProductHandlers(productProcessor ProductProcessor) *ProductHandlers {
	return &ProductHandlers{productProcessor: productProcessor}
}

var allowedProductTypes = map[string]bool{
	"электроника": true,
	"одежда":      true,
	"обувь":       true,
}

func (h *ProductHandlers) AddProductHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Type  string `json:"type"`
			PvzId string `json:"pvzId"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request body"})
		}

		if _, err := uuid.Parse(body.PvzId); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid pvzId format"})
		}

		if !allowedProductTypes[body.Type] {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid product type"})
		}

		product, err := h.productProcessor.AddProduct(body.PvzId, body.Type)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: err.Error()})
		}

		prometheus.ProductsAdded.Inc()
		return c.Status(fiber.StatusCreated).JSON(product)
	}
}

func (h *ProductHandlers) DeleteLastProductHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		pvzId := c.Params("pvzId")

		if _, err := uuid.Parse(pvzId); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid pvzId format"})
		}

		if err := h.productProcessor.DeleteLastProduct(pvzId); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: err.Error()})
		}

		return c.SendStatus(fiber.StatusOK)
	}
}
