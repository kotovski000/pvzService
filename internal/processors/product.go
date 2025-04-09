package processors

import (
	"errors"
	"github.com/google/uuid"
	"pvzService/internal/models"
)

func (p *Processor) AddProduct(productType string, pvzId string) (models.Product, error) {
	product := models.Product{}

	if productType != "электроника" && productType != "одежда" && productType != "обувь" {
		return product, errors.New("Invalid product type")
	}

	receptionID, err := p.Repo.GetOpenReceptionID(pvzId)

	if err != nil {
		return product, errors.New("No open reception for this PVZ")
	}

	productID := uuid.New().String()
	err = p.Repo.AddProduct(productID, receptionID, productType)

	if err != nil {
		return product, errors.New("Failed to add product")
	}

	product, err = p.Repo.GetProduct(productID)

	if err != nil {
		return product, errors.New("Failed to retrieve product")
	}

	return product, nil
}

func (p *Processor) DeleteLastProduct(pvzId string) error {
	err := p.Repo.DeleteLastProduct(pvzId)

	if err != nil {
		return errors.New("Failed to delete product")
	}

	return nil
}
