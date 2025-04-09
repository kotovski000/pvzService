package repository

import (
	"fmt"
	"pvzService/internal/models"
	"time"
)

func (r *Repository) CreateReception(pvzId string) (models.Reception, error) {
	var openReceptionExists bool
	err := r.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = 'in_progress')", pvzId).Scan(&openReceptionExists)

	if err != nil {
		return models.Reception{}, err
	}

	if openReceptionExists {
		return models.Reception{}, fmt.Errorf("Open reception already exists for this PVZ")
	}

	receptionID := fmt.Sprintf("%v", time.Now().Unix())
	_, err = r.DB.Exec("INSERT INTO receptions (id, pvz_id, status) VALUES ($1, $2, $3)", receptionID, pvzId, "in_progress")

	if err != nil {
		return models.Reception{}, err
	}

	var reception models.Reception
	err = r.DB.QueryRow("SELECT id, created_at, pvz_id, status FROM receptions WHERE id = $1", receptionID).Scan(&reception.ID, &reception.DateTime, &reception.PvzId, &reception.Status)

	return reception, err
}
func (r *Repository) CloseLastReception(pvzId string) (models.Reception, error) {
	var reception models.Reception
	err := r.DB.QueryRow("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'", pvzId).Scan(&reception.ID, &reception.DateTime, &reception.PvzId, &reception.Status, &reception.ClosedAt)

	if err != nil {
		return reception, err
	}

	now := time.Now()
	_, err = r.DB.Exec("UPDATE receptions SET status = 'close', closed_at = $1 WHERE id = $2", now, reception.ID)

	if err != nil {
		return reception, err
	}

	reception.Status = "close"
	reception.ClosedAt = &now

	return reception, nil
}

func (r *Repository) GetOpenReceptionID(pvzId string) (string, error) {
	var receptionID string
	err := r.DB.QueryRow("SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'", pvzId).Scan(&receptionID)
	return receptionID, err
}
