package services

import (
	"context"
	"strconv"
	"strings"

	"odontoestoque/internal/models"
)

const (
	defaultMovementLimit = 100
	maxMovementLimit     = 500
)

type MovementRepository interface {
	FindAll(ctx context.Context, f models.MovementFilter) ([]models.Movement, error)
}

type MovementService struct {
	repo MovementRepository
}

func NewMovementService(repo MovementRepository) *MovementService {
	return &MovementService{repo: repo}
}

// List valida os filtros vindos da query string e busca o histórico.
// tipo: "" ou "todas" = todos; limite: "" = 100, máximo 500.
func (s *MovementService) List(ctx context.Context, tipo, limite string) ([]models.Movement, error) {
	fields := map[string]string{}
	f := models.MovementFilter{Limite: defaultMovementLimit}

	switch t := strings.ToLower(strings.TrimSpace(tipo)); t {
	case "", "todas", "todos":
	case "entrada", "saida", "ajuste":
		f.Tipo = &t
	default:
		fields["tipo"] = "use entrada, saida ou ajuste"
	}

	if limite != "" {
		n, err := strconv.Atoi(limite)
		if err != nil || n < 1 || n > maxMovementLimit {
			fields["limite"] = "deve ser um número de 1 a 500"
		} else {
			f.Limite = n
		}
	}

	if len(fields) > 0 {
		return nil, &models.ValidationError{Fields: fields}
	}
	return s.repo.FindAll(ctx, f)
}
