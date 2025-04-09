package repository

import (
	"pvzService/internal/models"
)

func (r *Repository) AddProduct(productID string, receptionID string, productType string) error {
	_, err := r.DB.Exec("INSERT INTO products (id, reception_id, type) VALUES ($1, $2, $3)", productID, receptionID, productType)
	return err
}

func (r *Repository) GetProduct(productID string) (models.Product, error) {
	var product models.Product
	err := r.DB.QueryRow("SELECT id, created_at, type, reception_id FROM products WHERE id = $1", productID).Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionId)
	return product, err
}

func (r *Repository) DeleteLastProduct(pvzId string) error {
	var receptionID string
	err := r.DB.QueryRow("SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'", pvzId).Scan(&receptionID)

	if err != nil {
		return err
	}

	var productID string
	err = r.DB.QueryRow("SELECT id FROM products WHERE reception_id = $1 ORDER BY created_at DESC LIMIT 1", receptionID).Scan(&productID)

	if err != nil {
		return err
	}

	_, err = r.DB.Exec("DELETE FROM products WHERE id = $1", productID)

	return err
}
