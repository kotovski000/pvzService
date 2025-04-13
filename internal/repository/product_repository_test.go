package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvzService/internal/models"
)

func TestProductRepository_AddProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	receptionID := uuid.NewString()
	productID := uuid.NewString()

	mock.ExpectExec("INSERT INTO products").
		WithArgs(productID, receptionID, "электроника").
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := repo.AddProduct(receptionID, "электроника", func() uuid.UUID {
		return uuid.MustParse(productID)
	})

	assert.NoError(t, err)
	assert.Equal(t, productID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetProductByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.NewString()
	expected := models.Product{
		ID:   productID,
		Type: "электроника",
	}

	mock.ExpectQuery("SELECT id, created_at, type, reception_id FROM products").
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "type", "reception_id"}).
			AddRow(expected.ID, expected.DateTime, expected.Type, expected.ReceptionId))

	product, err := repo.GetProductByID(productID)
	assert.NoError(t, err)
	assert.Equal(t, expected, product)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_DeleteProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.NewString()

	mock.ExpectExec("DELETE FROM products").
		WithArgs(productID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteProduct(productID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
