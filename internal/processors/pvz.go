package processors

import (
	"errors"
	"github.com/google/uuid"
	"pvzService/internal/models"
)

func (p *Processor) CreatePVZ(city string) (models.PVZ, error) {
	pvz := models.PVZ{}
	if city != "Москва" && city != "Санкт-Петербург" && city != "Казань" {
		return pvz, errors.New("Invalid city")
	}

	pvzID := uuid.New().String()
	err := p.Repo.CreatePVZ(pvzID, city)

	if err != nil {
		return pvz, errors.New("Failed to create PVZ")
	}

	pvz, err = p.Repo.GetPVZ(pvzID)

	if err != nil {
		return pvz, errors.New("Failed to retrieve PVZ after creation")
	}

	return pvz, nil
}

func (p *Processor) GetPVZList(startDate string, endDate string, page string, limit string) (interface{}, error) {
	pvzs, err := p.Repo.GetPVZList(startDate, endDate, page, limit)

	if err != nil {
		return nil, errors.New("Failed to retrieve PVZ list")
	}

	return pvzs, nil
}
