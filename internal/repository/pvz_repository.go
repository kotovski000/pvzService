package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"pvzService/internal/models"
	"pvzService/internal/utils"
)

type PVZRepository interface {
	CreatePVZ(city string, idGenerator func() uuid.UUID) (models.PVZ, error)
	GetPVZByID(id string) (models.PVZ, error)
	ListPVZsWithRelations(startDate, endDate time.Time, limit, offset int) ([]PVZResponse, error)
}

type PVZRepositoryImpl struct {
	db *sql.DB
}

func NewPVZRepository(db *sql.DB) *PVZRepositoryImpl {
	return &PVZRepositoryImpl{db: db}
}

func (r *PVZRepositoryImpl) CreatePVZ(city string, idGenerator func() uuid.UUID) (models.PVZ, error) {
	pvzID := idGenerator().String()
	_, err := r.db.Exec("INSERT INTO pvz (id, city) VALUES ($1, $2)", pvzID, city)
	if err != nil {
		return models.PVZ{}, err
	}

	var pvz models.PVZ
	err = r.db.QueryRow("SELECT id, registration_date, city FROM pvz WHERE id = $1", pvzID).
		Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	return pvz, err
}

func (r *PVZRepositoryImpl) GetPVZByID(id string) (models.PVZ, error) {
	var pvz models.PVZ
	err := r.db.QueryRow("SELECT id, registration_date, city FROM pvz WHERE id = $1", id).
		Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	return pvz, err
}

type PVZResponse struct {
	PVZ        models.PVZ          `json:"pvz"`
	Receptions []ReceptionResponse `json:"receptions"`
}

type ReceptionResponse struct {
	Reception models.Reception `json:"reception"`
	Products  []models.Product `json:"products"`
}

func (r *PVZRepositoryImpl) ListPVZsWithRelations(startDate, endDate time.Time, limit, offset int) ([]PVZResponse, error) {
	var rows *sql.Rows
	var err error

	query := `
        SELECT 
            p.id, p.registration_date, p.city,
            r.id, r.created_at, r.pvz_id, r.status, r.closed_at,
            pr.id, pr.created_at, pr.type, pr.reception_id
        FROM pvz p
        LEFT JOIN receptions r ON p.id = r.pvz_id
        LEFT JOIN products pr ON r.id = pr.reception_id
    `

	if !startDate.IsZero() && !endDate.IsZero() {
		query += " WHERE p.registration_date >= $1 AND p.registration_date <= $2"
		query += " ORDER BY p.registration_date"
		query += " LIMIT $3 OFFSET $4"
		rows, err = r.db.Query(query, startDate, endDate, limit, offset)
	} else {
		query += " ORDER BY p.registration_date"
		query += " LIMIT $1 OFFSET $2"
		rows, err = r.db.Query(query, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pvzMap := make(map[string]*PVZResponse)
	receptionMap := make(map[string]*struct {
		reception models.Reception
		products  []models.Product
	})

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

		if err := rows.Scan(
			&pvzID, &pvzRegDate, &pvzCity,
			&receptionID, &receptionCreatedAt, &receptionPvzID, &receptionStatus, &receptionClosedAt,
			&productID, &productCreatedAt, &productType, &productReceptionID,
		); err != nil {
			return nil, err
		}

		if pvzID.Valid {
			if _, exists := pvzMap[pvzID.String]; !exists {
				pvzMap[pvzID.String] = &PVZResponse{
					PVZ: models.PVZ{
						ID:               pvzID.String,
						RegistrationDate: pvzRegDate.Time,
						City:             pvzCity.String,
					},
					Receptions: []ReceptionResponse{},
				}
			}
		}

		if receptionID.Valid && pvzID.Valid {
			receptionKey := receptionID.String
			if _, exists := receptionMap[receptionKey]; !exists {
				receptionMap[receptionKey] = &struct {
					reception models.Reception
					products  []models.Product
				}{
					reception: models.Reception{
						ID:       receptionID.String,
						DateTime: receptionCreatedAt.Time,
						PvzId:    receptionPvzID.String,
						Status:   receptionStatus.String,
						ClosedAt: utils.NullableTime(receptionClosedAt),
					},
					products: []models.Product{},
				}
			}
		}

		if productID.Valid && productReceptionID.Valid {
			if reception, exists := receptionMap[productReceptionID.String]; exists {
				reception.products = append(reception.products, models.Product{
					ID:          productID.String,
					DateTime:    productCreatedAt.Time,
					Type:        productType.String,
					ReceptionId: productReceptionID.String,
				})
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, reception := range receptionMap {
		if pvz, exists := pvzMap[reception.reception.PvzId]; exists {
			pvz.Receptions = append(pvz.Receptions, ReceptionResponse{
				Reception: reception.reception,
				Products:  reception.products,
			})
		}
	}

	result := make([]PVZResponse, 0, len(pvzMap))
	for _, pvz := range pvzMap {
		result = append(result, *pvz)
	}

	return result, nil
}
