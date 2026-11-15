package services

import (
	"context"

	"odontoestoque/internal/models"
)

type LanguageRepository interface {
	FindAll(ctx context.Context) ([]models.Language, error)
}

type LanguageService struct {
	repo LanguageRepository
}

func NewLanguageService(repo LanguageRepository) *LanguageService {
	return &LanguageService{repo: repo}
}

func (s *LanguageService) List(ctx context.Context) ([]models.Language, error) {
	return s.repo.FindAll(ctx)
}
