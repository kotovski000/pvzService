package processors

import (
	"database/sql"
	"errors"
	"time"

	"pvzService/internal/models"
	"pvzService/internal/repository"
)

type ReceptionProcessor struct {
	receptionRepo *repository.ReceptionRepository
}

func NewReceptionProcessor(receptionRepo *repository.ReceptionRepository) *ReceptionProcessor {
	return &ReceptionProcessor{receptionRepo: receptionRepo}
}

func (p *ReceptionProcessor) CreateReception(pvzID string) (models.Reception, error) {
	hasOpen, err := p.receptionRepo.HasOpenReception(pvzID)
	if err != nil {
		return models.Reception{}, errors.New("database error")
	}
	if hasOpen {
		return models.Reception{}, errors.New("open reception already exists for this PVZ")
	}

	receptionID, err := p.receptionRepo.CreateReception(pvzID)
	if err != nil {
		return models.Reception{}, errors.New("failed to create reception")
	}

	return p.receptionRepo.GetReceptionByID(receptionID)
}

func (p *ReceptionProcessor) CloseLastReception(pvzID string) (models.Reception, error) {
	reception, err := p.receptionRepo.GetOpenReception(pvzID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Reception{}, errors.New("no open reception found for this PVZ")
		}
		return models.Reception{}, errors.New("database error")
	}

	now := time.Now()
	if err := p.receptionRepo.CloseReception(reception.ID, now); err != nil {
		return models.Reception{}, errors.New("failed to close reception")
	}

	reception.Status = "close"
	reception.ClosedAt = &now
	return reception, nil
}
