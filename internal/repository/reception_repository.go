package repository

import (
	"database/sql"
	"github.com/google/uuid"
	"time"

	"pvzService/internal/models"
)

type ReceptionRepository struct {
	db *sql.DB
}

func NewReceptionRepository(db *sql.DB) *ReceptionRepository {
	return &ReceptionRepository{db: db}
}

func (r *ReceptionRepository) CreateReception(pvzID string) (string, error) {
	receptionID := uuid.New().String()
	_, err := r.db.Exec("INSERT INTO receptions (id, pvz_id, status) VALUES ($1, $2, $3)",
		receptionID, pvzID, "in_progress")
	return receptionID, err
}

func (r *ReceptionRepository) GetReceptionByID(id string) (models.Reception, error) {
	var reception models.Reception
	err := r.db.QueryRow("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE id = $1", id).
		Scan(&reception.ID, &reception.DateTime, &reception.PvzId, &reception.Status, &reception.ClosedAt)
	return reception, err
}

func (r *ReceptionRepository) GetOpenReception(pvzID string) (models.Reception, error) {
	var reception models.Reception
	err := r.db.QueryRow(
		"SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'",
		pvzID).
		Scan(&reception.ID, &reception.DateTime, &reception.PvzId, &reception.Status, &reception.ClosedAt)
	return reception, err
}

func (r *ReceptionRepository) CloseReception(id string, closeTime time.Time) error {
	_, err := r.db.Exec("UPDATE receptions SET status = 'close', closed_at = $1 WHERE id = $2",
		closeTime, id)
	return err
}

func (r *ReceptionRepository) HasOpenReception(pvzID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = 'in_progress')",
		pvzID).
		Scan(&exists)
	return exists, err
}
