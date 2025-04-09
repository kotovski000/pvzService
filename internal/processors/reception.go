package processors

import (
	"errors"
	"pvzService/internal/models"
)

func (p *Processor) CreateReception(pvzId string) (models.Reception, error) {
	reception, err := p.Repo.CreateReception(pvzId)

	if err != nil {
		return reception, errors.New("Failed to create reception")
	}

	return reception, nil
}

func (p *Processor) CloseLastReception(pvzId string) (models.Reception, error) {
	reception, err := p.Repo.CloseLastReception(pvzId)

	if err != nil {
		return reception, errors.New("Failed to close reception")
	}

	return reception, nil
}
