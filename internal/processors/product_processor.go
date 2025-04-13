package processors

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"

	"pvzService/internal/models"
)

type ProductRepository interface {
	AddProduct(receptionID string, productType string, idGenerator func() uuid.UUID) (string, error)
	GetProductByID(id string) (models.Product, error)
	GetLastProduct(receptionID string) (models.Product, error)
	DeleteProduct(id string) error
}

type ReceptionRepository interface {
	GetOpenReception(pvzID string) (models.Reception, error)
}

type ProductProcessor struct {
	productRepo   ProductRepository
	receptionRepo ReceptionRepository
}

func NewProductProcessor(
	productRepo ProductRepository,
	receptionRepo ReceptionRepository,
) *ProductProcessor {
	return &ProductProcessor{
		productRepo:   productRepo,
		receptionRepo: receptionRepo,
	}
}

func (p *ProductProcessor) AddProduct(pvzID, productType string) (models.Product, error) {
	allowedTypes := map[string]bool{
		"электроника": true,
		"одежда":      true,
		"обувь":       true,
	}

	if !allowedTypes[productType] {
		return models.Product{}, errors.New("invalid product type")
	}

	reception, err := p.receptionRepo.GetOpenReception(pvzID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Product{}, errors.New("no open reception for this PVZ")
		}
		return models.Product{}, errors.New("database error")
	}

	productID, err := p.productRepo.AddProduct(reception.ID, productType, uuid.New)
	if err != nil {
		return models.Product{}, errors.New("failed to add product")
	}

	return p.productRepo.GetProductByID(productID)
}

func (p *ProductProcessor) DeleteLastProduct(pvzID string) error {
	reception, err := p.receptionRepo.GetOpenReception(pvzID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no open reception for this PVZ")
		}
		return errors.New("database error")
	}

	product, err := p.productRepo.GetLastProduct(reception.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no products to delete in this reception")
		}
		return errors.New("database error")
	}

	return p.productRepo.DeleteProduct(product.ID)
}
