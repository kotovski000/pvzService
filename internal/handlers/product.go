package handlers

import (
	"database/sql"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"pvzService/internal/models"
)

func AddProductHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Type  string `json:"type"`
			PvzId string `json:"pvzId"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request"})
		}

		if body.Type != "электроника" && body.Type != "одежда" && body.Type != "обувь" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid product type"})
		}

		var receptionID string
		err := db.QueryRow("SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'", body.PvzId).Scan(&receptionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "No open reception for this PVZ"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Database error"})
		}

		productID := uuid.New().String()
		_, err = db.Exec("INSERT INTO products (id, reception_id, type) VALUES ($1, $2, $3)", productID, receptionID, body.Type)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to add product"})
		}

		var product models.Product
		err = db.QueryRow("SELECT id, created_at, type, reception_id FROM products WHERE id = $1", productID).Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to retrieve product"})
		}

		return c.Status(fiber.StatusCreated).JSON(product)
	}
}

func DeleteLastProductHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pvzId := c.Params("pvzId")

		_, err := uuid.Parse(pvzId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid pvzId format"})
		}

		var receptionID string
		err = db.QueryRow("SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'", pvzId).Scan(&receptionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "No open reception for this PVZ"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Database error"})
		}

		var productID string
		err = db.QueryRow("SELECT id FROM products WHERE reception_id = $1 ORDER BY created_at DESC LIMIT 1", receptionID).Scan(&productID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "No products to delete in this reception"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Database error"})
		}

		_, err = db.Exec("DELETE FROM products WHERE id = $1", productID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to delete product"})
		}

		return c.SendStatus(fiber.StatusOK)
	}
}
