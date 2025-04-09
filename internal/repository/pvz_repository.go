package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"pvzService/internal/models"
	"pvzService/internal/utils"
)

type PVZRepository struct {
	db *sql.DB
}

func NewPVZRepository(db *sql.DB) *PVZRepository {
	return &PVZRepository{db: db}
}

func (r *PVZRepository) CreatePVZ(city string) (models.PVZ, error) {
	pvzID := uuid.New().String()
	_, err := r.db.Exec("INSERT INTO pvz (id, city) VALUES ($1, $2)", pvzID, city)
	if err != nil {
		return models.PVZ{}, err
	}

	var pvz models.PVZ
	err = r.db.QueryRow("SELECT id, registration_date, city FROM pvz WHERE id = $1", pvzID).
		Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	return pvz, err
}

func (r *PVZRepository) GetPVZByID(id string) (models.PVZ, error) {
	var pvz models.PVZ
	err := r.db.QueryRow("SELECT id, registration_date, city FROM pvz WHERE id = $1", id).
		Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	return pvz, err
}

type PVZWithRelations struct {
	PVZ        models.PVZ
	Receptions []ReceptionWithProducts
}

type ReceptionWithProducts struct {
	Reception models.Reception
	Products  []models.Product
}

func (r *PVZRepository) ListPVZsWithRelations(startDate, endDate time.Time, limit, offset int) ([]PVZWithRelations, error) {
	var rows *sql.Rows
	var err error

	baseQuery := `
        SELECT 
            p.id, p.registration_date, p.city,
            r.id, r.created_at, r.pvz_id, r.status, r.closed_at,
            pr.id, pr.created_at, pr.type, pr.reception_id
        FROM pvz p
        LEFT JOIN receptions r ON p.id = r.pvz_id
        LEFT JOIN products pr ON r.id = pr.reception_id
    `

	if !startDate.IsZero() && !endDate.IsZero() {
		baseQuery += " WHERE p.registration_date >= $1 AND p.registration_date <= $2"
		baseQuery += " ORDER BY p.registration_date"
		baseQuery += " LIMIT $3 OFFSET $4"
		rows, err = r.db.Query(baseQuery, startDate, endDate, limit, offset)
	} else {
		baseQuery += " ORDER BY p.registration_date"
		baseQuery += " LIMIT $1 OFFSET $2"
		rows, err = r.db.Query(baseQuery, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pvzMap := make(map[string]*PVZWithRelations)
	receptionMap := make(map[string]*models.Reception)
	productMap := make(map[string][]models.Product)

	for rows.Next() {
		var (
			pvzID, receptionID, productID sql.NullString
			pvzRegDate                    sql.NullTime
			pvzCity                       sql.NullString
			receptionCreatedAt            sql.NullTime
			receptionPvzID                sql.NullString
			receptionStatus               sql.NullString
			receptionClosedAt             sql.NullTime
			productCreatedAt              sql.NullTime
			productType                   sql.NullString
			productReceptionID            sql.NullString
		)

		err = rows.Scan(
			&pvzID, &pvzRegDate, &pvzCity,
			&receptionID, &receptionCreatedAt, &receptionPvzID, &receptionStatus, &receptionClosedAt,
			&productID, &productCreatedAt, &productType, &productReceptionID,
		)
		if err != nil {
			return nil, err
		}

		// Обработка PVZ
		if pvzID.Valid {
			if _, exists := pvzMap[pvzID.String]; !exists {
				pvzMap[pvzID.String] = &PVZWithRelations{
					PVZ: models.PVZ{
						ID:               pvzID.String,
						RegistrationDate: pvzRegDate.Time,
						City:             pvzCity.String,
					},
					Receptions: []ReceptionWithProducts{},
				}
			}
		}

		// Обработка Reception
		if receptionID.Valid {
			if _, exists := receptionMap[receptionID.String]; !exists {
				reception := models.Reception{
					ID:       receptionID.String,
					DateTime: receptionCreatedAt.Time,
					PvzId:    receptionPvzID.String,
					Status:   receptionStatus.String,
					ClosedAt: utils.NullableTime(receptionClosedAt),
				}
				receptionMap[receptionID.String] = &reception
			}
		}

		// Обработка Product
		if productID.Valid && productReceptionID.Valid {
			product := models.Product{
				ID:          productID.String,
				DateTime:    productCreatedAt.Time,
				Type:        productType.String,
				ReceptionId: productReceptionID.String,
			}
			productMap[productReceptionID.String] = append(productMap[productReceptionID.String], product)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Собираем окончательную структуру
	for _, pvz := range pvzMap {
		for receptionID, reception := range receptionMap {
			if reception.PvzId == pvz.PVZ.ID {
				receptionWithProducts := ReceptionWithProducts{
					Reception: *reception,
					Products:  productMap[receptionID],
				}
				pvz.Receptions = append(pvz.Receptions, receptionWithProducts)
			}
		}
	}

	// Преобразуем map в slice
	result := make([]PVZWithRelations, 0, len(pvzMap))
	for _, pvz := range pvzMap {
		result = append(result, *pvz)
	}

	return result, nil
}
