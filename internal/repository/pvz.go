package repository

import (
	"pvzService/internal/models"
)

func (r *Repository) CreatePVZ(pvzID string, city string) error {
	_, err := r.DB.Exec("INSERT INTO pvz (id, city) VALUES ($1, $2)", pvzID, city)
	return err
}

func (r *Repository) GetPVZ(pvzID string) (models.PVZ, error) {
	var pvz models.PVZ
	err := r.DB.QueryRow("SELECT id, registration_date, city FROM pvz WHERE id = $1", pvzID).Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	return pvz, err
}

func (r *Repository) GetPVZList(startDate string, endDate string, page string, limit string) (interface{}, error) {
	return nil, nil
}
