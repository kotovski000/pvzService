package handlers

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"pvzService/internal/models"
	"pvzService/internal/utils"
)

func CreatePVZHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body models.PVZ

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid request"})
		}

		if body.City != "Москва" && body.City != "Санкт-Петербург" && body.City != "Казань" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid city"})
		}

		pvzID := uuid.New().String()
		_, err := db.Exec("INSERT INTO pvz (id, city) VALUES ($1, $2)", pvzID, body.City)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to create PVZ"})
		}

		var pvz models.PVZ
		err = db.QueryRow("SELECT id, registration_date, city FROM pvz WHERE id = $1", pvzID).Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to retrieve PVZ after creation"})
		}

		return c.Status(fiber.StatusCreated).JSON(pvz)
	}
}

func GetPVZListHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startDateStr := c.Query("startDate")
		endDateStr := c.Query("endDate")
		pageStr := c.Query("page", "1")
		limitStr := c.Query("limit", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid page number"})
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 30 {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid limit"})
		}

		var startDate, endDate time.Time
		if startDateStr != "" {
			startDate, err = time.Parse(time.RFC3339, startDateStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid start date format"})
			}
		}

		if endDateStr != "" {
			endDate, err = time.Parse(time.RFC3339, endDateStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{Message: "Invalid end date format"})
			}
		}

		offset := (page - 1) * limit

		var rows *sql.Rows
		if startDateStr != "" && endDateStr != "" {
			rows, err = db.Query(`
			SELECT 
				p.id, p.registration_date, p.city,
				r.id, r.created_at, r.pvz_id, r.status, r.closed_at,
				pr.id, pr.created_at, pr.type, pr.reception_id
			FROM pvz p
			LEFT JOIN receptions r ON p.id = r.pvz_id
			LEFT JOIN products pr ON r.id = pr.reception_id
			WHERE p.registration_date >= $1 AND p.registration_date <= $2
			ORDER BY p.registration_date
			LIMIT $3 OFFSET $4
		`, startDate, endDate, limit, offset)
		} else {
			rows, err = db.Query(`
				SELECT 
					p.id, p.registration_date, p.city,
					r.id, r.created_at, r.pvz_id, r.status, r.closed_at,
					pr.id, pr.created_at, pr.type, pr.reception_id
				FROM pvz p
				LEFT JOIN receptions r ON p.id = r.pvz_id
				LEFT JOIN products pr ON r.id = pr.reception_id
				ORDER BY p.registration_date
				LIMIT $1 OFFSET $2
			`, limit, offset)
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.Error{Message: "Failed to retrieve PVZ list"})
		}
		defer rows.Close()

		pvzMap := make(map[string]models.PVZ)
		receptionMap := make(map[string]models.Reception)
		productMap := make(map[string]models.Product)

		var pvzList []map[string]interface{}

		for rows.Next() {
			var pvzID, receptionID, productID sql.NullString
			var pvzRegDate sql.NullTime
			var pvzCity sql.NullString
			var receptionCreatedAt sql.NullTime
			var receptionPvzID sql.NullString
			var receptionStatus sql.NullString
			var receptionClosedAt sql.NullTime
			var productCreatedAt sql.NullTime
			var productType sql.NullString
			var productReceptionID sql.NullString
			err = rows.Scan(
				&pvzID, &pvzRegDate, &pvzCity,
				&receptionID, &receptionCreatedAt, &receptionPvzID, &receptionStatus, &receptionClosedAt,
				&productID, &productCreatedAt, &productType, &productReceptionID,
			)

			if err != nil {
				log.Println("Error scanning row:", err)
				continue
			}

			if pvzID.Valid {
				if _, ok := pvzMap[pvzID.String]; !ok {
					pvzMap[pvzID.String] = models.PVZ{
						ID:               pvzID.String,
						RegistrationDate: pvzRegDate.Time,
						City:             pvzCity.String,
					}
				}
			}

			if receptionID.Valid {
				if _, ok := receptionMap[receptionID.String]; !ok {
					receptionMap[receptionID.String] = models.Reception{
						ID:       receptionID.String,
						DateTime: receptionCreatedAt.Time,
						PvzId:    receptionPvzID.String,
						Status:   receptionStatus.String,
						ClosedAt: utils.NullableTime(receptionClosedAt),
					}
				}
			}

			if productID.Valid {
				productMap[productID.String] = models.Product{
					ID:          productID.String,
					DateTime:    productCreatedAt.Time,
					Type:        productType.String,
					ReceptionId: productReceptionID.String,
				}
			}
		}

		if err := rows.Err(); err != nil {
			log.Println("Error iterating rows:", err)
		}

		pvzWithReceptions := make(map[string]map[string]interface{})

		for _, pvz := range pvzMap {
			pvzWithReceptions[pvz.ID] = map[string]interface{}{
				"pvz":        pvz,
				"receptions": make(map[string]map[string]interface{}),
			}
		}

		for _, reception := range receptionMap {
			if pvzEntry, ok := pvzWithReceptions[reception.PvzId]; ok {
				receptionMap := pvzEntry["receptions"].(map[string]map[string]interface{})
				receptionMap[reception.ID] = map[string]interface{}{
					"reception": reception,
					"products":  make([]models.Product, 0),
				}
			}
		}

		for _, product := range productMap {
			for _, pvzData := range pvzWithReceptions {
				receptionsMap := pvzData["receptions"].(map[string]map[string]interface{})
				for receptionID, receptionData := range receptionsMap {
					reception := receptionData["reception"].(models.Reception)
					if product.ReceptionId == receptionID && product.ReceptionId == reception.ID {
						products := receptionData["products"].([]models.Product)
						receptionData["products"] = append(products, product)
						receptionsMap[receptionID] = receptionData
						pvzData["receptions"] = receptionsMap
						pvzWithReceptions[pvzData["pvz"].(models.PVZ).ID] = pvzData
						break
					}
				}
			}
		}

		for _, pvz := range pvzWithReceptions {
			var receptionsList []map[string]interface{}
			for _, receptionData := range pvz["receptions"].(map[string]map[string]interface{}) {
				receptionsList = append(receptionsList, receptionData)
			}
			pvz["receptions"] = receptionsList
			pvzList = append(pvzList, pvz)
		}

		return c.JSON(pvzList)
	}
}
