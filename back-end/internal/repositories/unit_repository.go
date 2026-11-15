package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"odontoestoque/infra/database/queries"
	"odontoestoque/internal/models"
)

type UnitRepository struct {
	db *pgxpool.Pool
}

func NewUnitRepository(db *pgxpool.Pool) *UnitRepository {
	return &UnitRepository{db: db}
}

// FindAll executa queries/unidades_medida/select_all.sql.
func (r *UnitRepository) FindAll(ctx context.Context) ([]models.Unit, error) {
	rows, err := r.db.Query(ctx, queries.UnidadeMedidaSelectAll)
	if err != nil {
		return nil, fmt.Errorf("unit repository: find all: %w", err)
	}
	defer rows.Close()

	units := []models.Unit{}
	for rows.Next() {
		var u models.Unit
		if err := rows.Scan(&u.ID, &u.Nome, &u.CriadoEm); err != nil {
			return nil, fmt.Errorf("unit repository: find all: %w", err)
		}
		units = append(units, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unit repository: find all: %w", err)
	}
	return units, nil
}
