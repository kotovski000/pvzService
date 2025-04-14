package repository

import (
	"database/sql"
	"github.com/google/uuid"
	"time"

	"pvzService/internal/models"
)

type ReceptionRepository interface {
	CreateReception(pvzID string, idGenerator func() uuid.UUID) (string, error)
	GetReceptionByID(id string) (models.Reception, error)
	GetOpenReception(pvzID string) (models.Reception, error)
	CloseReception(id string, closeTime time.Time) error
	HasOpenReception(pvzID string) (bool, error)
}

type ReceptionRepositoryImpl struct {
	db *sql.DB
}

func NewReceptionRepository(db *sql.DB) *ReceptionRepositoryImpl {
	return &ReceptionRepositoryImpl{db: db}
}

func (r *ReceptionRepositoryImpl) CreateReception(pvzID string, idGenerator func() uuid.UUID) (string, error) {
	receptionID := idGenerator().String()
	_, err := r.db.Exec("INSERT INTO receptions (id, pvz_id, status, created_at) VALUES ($1, $2, $3, $4)",
		receptionID, pvzID, "in_progress", time.Now())
	return receptionID, err
}

func (r *ReceptionRepositoryImpl) GetReceptionByID(id string) (models.Reception, error) {
	var reception models.Reception
	err := r.db.QueryRow("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE id = $1", id).
		Scan(&reception.ID, &reception.DateTime, &reception.PvzId, &reception.Status, &reception.ClosedAt)
	return reception, err
}

func (r *ReceptionRepositoryImpl) GetOpenReception(pvzID string) (models.Reception, error) {
	var reception models.Reception
	err := r.db.QueryRow(
		"SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'",
		pvzID).
		Scan(&reception.ID, &reception.DateTime, &reception.PvzId, &reception.Status, &reception.ClosedAt)
	return reception, err
}

func (r *ReceptionRepositoryImpl) CloseReception(id string, closeTime time.Time) error {
	_, err := r.db.Exec("UPDATE receptions SET status = 'close', closed_at = $1 WHERE id = $2",
		closeTime, id)
	return err
}

func (r *ReceptionRepositoryImpl) HasOpenReception(pvzID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = 'in_progress')",
		pvzID).
		Scan(&exists)
	return exists, err
}
