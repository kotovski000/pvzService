package handlers

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"pvzService/internal/models"
)

func CreateReceptionHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			PvzId string `json:"pvzId"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request"})
		}

		_, err := uuid.Parse(body.PvzId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid pvzId format"})
		}

		var openReceptionExists bool
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = 'in_progress')", body.PvzId).Scan(&openReceptionExists)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Database error"})
		}
		if openReceptionExists {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Open reception already exists for this PVZ"})
		}

		receptionID := uuid.New().String()
		_, err = db.Exec("INSERT INTO receptions (id, pvz_id, status) VALUES ($1, $2, $3)", receptionID, body.PvzId, "in_progress")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to create reception"})
		}

		var reception models.Reception
		err = db.QueryRow("SELECT id, created_at, pvz_id, status FROM receptions WHERE id = $1", receptionID).Scan(&reception.ID, &reception.DateTime, &reception.PvzId, &reception.Status)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to retrieve reception"})
		}

		return c.Status(fiber.StatusCreated).JSON(reception)
	}
}

func CloseLastReceptionHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pvzId := c.Params("pvzId")

		_, err := uuid.Parse(pvzId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid pvzId format"})
		}

		var reception models.Reception
		err = db.QueryRow("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'", pvzId).Scan(&reception.ID, &reception.DateTime, &reception.PvzId, &reception.Status, &reception.ClosedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "No open reception found for this PVZ"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Database error"})
		}

		now := time.Now()
		_, err = db.Exec("UPDATE receptions SET status = 'close', closed_at = $1 WHERE id = $2", now, reception.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to close reception"})
		}

		reception.Status = "close"
		reception.ClosedAt = &now

		return c.JSON(reception)
	}
}
