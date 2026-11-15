package services

import (
	"context"

	"odontoestoque/internal/models"
)

type UnitRepository interface {
	FindAll(ctx context.Context) ([]models.Unit, error)
}

type UnitService struct {
	repo UnitRepository
}

func NewUnitService(repo UnitRepository) *UnitService {
	return &UnitService{repo: repo}
}

func (s *UnitService) List(ctx context.Context) ([]models.Unit, error) {
	return s.repo.FindAll(ctx)
}
