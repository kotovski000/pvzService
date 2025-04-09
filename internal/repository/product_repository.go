package repository

import (
	"database/sql"
	"github.com/google/uuid"

	"pvzService/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) AddProduct(receptionID, productType string) (string, error) {
	productID := uuid.New().String()
	_, err := r.db.Exec("INSERT INTO products (id, reception_id, type) VALUES ($1, $2, $3)",
		productID, receptionID, productType)
	return productID, err
}

func (r *ProductRepository) GetProductByID(id string) (models.Product, error) {
	var product models.Product
	err := r.db.QueryRow("SELECT id, created_at, type, reception_id FROM products WHERE id = $1", id).
		Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionId)
	return product, err
}

func (r *ProductRepository) GetLastProduct(receptionID string) (models.Product, error) {
	var product models.Product
	err := r.db.QueryRow(
		"SELECT id, created_at, type, reception_id FROM products WHERE reception_id = $1 ORDER BY created_at DESC LIMIT 1",
		receptionID).
		Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionId)
	return product, err
}

func (r *ProductRepository) DeleteProduct(id string) error {
	_, err := r.db.Exec("DELETE FROM products WHERE id = $1", id)
	return err
}
