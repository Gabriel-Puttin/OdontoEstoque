package services

import (
	"context"
	"strings"
	"unicode/utf8"

	"odontoestoque/internal/models"
)

type CategoryRepository interface {
	FindAll(ctx context.Context) ([]models.Category, error)
	FindByID(ctx context.Context, id int16) (*models.Category, error)
	FindProducts(ctx context.Context, categoryID int16) ([]models.CategoryProduct, error)
	Create(ctx context.Context, nome string) (*models.Category, error)
	Update(ctx context.Context, id int16, nome string) (*models.Category, error)
	Delete(ctx context.Context, id int16) error
}

type CategoryService struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context) ([]models.Category, error) {
	return s.repo.FindAll(ctx)
}

// Products devolve a categoria e seus produtos ativos (404 se a categoria não existir).
func (s *CategoryService) Products(ctx context.Context, id int16) (*models.CategoryProductsResponse, error) {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	products, err := s.repo.FindProducts(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.CategoryProductsResponse{Categoria: *category, Produtos: products}, nil
}

func (s *CategoryService) Create(ctx context.Context, req models.CategoryRequest) (*models.Category, error) {
	nome, err := validateCategoryName(req.Nome)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, nome)
}

func (s *CategoryService) Rename(ctx context.Context, id int16, req models.CategoryRequest) (*models.Category, error) {
	nome, err := validateCategoryName(req.Nome)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, nome)
}

func (s *CategoryService) Delete(ctx context.Context, id int16) error {
	return s.repo.Delete(ctx, id)
}

func validateCategoryName(raw string) (string, error) {
	nome := strings.TrimSpace(raw)
	switch {
	case nome == "":
		return "", &models.ValidationError{Fields: map[string]string{"nome": "obrigatório"}}
	case utf8.RuneCountInString(nome) > 80:
		return "", &models.ValidationError{Fields: map[string]string{"nome": "máximo de 80 caracteres"}}
	}
	return nome, nil
}
