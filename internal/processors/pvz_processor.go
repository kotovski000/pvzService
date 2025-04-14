package processors

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"pvzService/internal/models"
	"pvzService/internal/repository"
)

type PVZProcessor interface {
	CreatePVZ(city string) (models.PVZ, error)
	GetPVZByID(id string) (models.PVZ, error)
	ListPVZsWithRelations(startDate, endDate string, page, limit int) ([]repository.PVZResponse, error)
}

type PVZProcessorImpl struct {
	pvzRepo repository.PVZRepository
}

func NewPVZProcessor(pvzRepo repository.PVZRepository) *PVZProcessorImpl {
	return &PVZProcessorImpl{pvzRepo: pvzRepo}
}

func (p *PVZProcessorImpl) CreatePVZ(city string) (models.PVZ, error) {
	allowedCities := map[string]bool{
		"Москва":          true,
		"Санкт-Петербург": true,
		"Казань":          true,
	}

	if !allowedCities[city] {
		return models.PVZ{}, errors.New("invalid city")
	}

	return p.pvzRepo.CreatePVZ(city, uuid.New)
}

func (p *PVZProcessorImpl) GetPVZByID(id string) (models.PVZ, error) {
	return p.pvzRepo.GetPVZByID(id)
}

func (p *PVZProcessorImpl) ListPVZsWithRelations(startDate, endDate string, page, limit int) ([]repository.PVZResponse, error) {
	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse(time.RFC3339, startDate)
		if err != nil {
			return nil, errors.New("invalid start date format")
		}
	}

	if endDate != "" {
		end, err = time.Parse(time.RFC3339, endDate)
		if err != nil {
			return nil, errors.New("invalid end date format")
		}
	}

	if page < 1 {
		return nil, errors.New("invalid page number")
	}

	if limit < 1 || limit > 30 {
		return nil, errors.New("invalid limit")
	}

	offset := (page - 1) * limit
	return p.pvzRepo.ListPVZsWithRelations(start, end, limit, offset)
}
