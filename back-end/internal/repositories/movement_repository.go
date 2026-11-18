package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"odontoestoque/infra/database/queries"
	"odontoestoque/internal/models"
)

type MovementRepository struct {
	db *pgxpool.Pool
}

func NewMovementRepository(db *pgxpool.Pool) *MovementRepository {
	return &MovementRepository{db: db}
}

// FindAll executa queries/movimentacoes/select_all.sql.
func (r *MovementRepository) FindAll(ctx context.Context, f models.MovementFilter) ([]models.Movement, error) {
	rows, err := r.db.Query(ctx, queries.MovimentacaoSelectAll, f.Tipo, f.Limite)
	if err != nil {
		return nil, fmt.Errorf("movement repository: find all: %w", err)
	}
	defer rows.Close()

	movements := []models.Movement{}
	for rows.Next() {
		var m models.Movement
		if err := rows.Scan(
			&m.ID, &m.Tipo, &m.ProdutoID, &m.Produto, &m.ProdutoExcluido,
			&m.QuantidadeAnterior, &m.QuantidadePosterior, &m.Delta, &m.UnidadeMedida,
			&m.UsuarioID, &m.Usuario, &m.CriadoEm,
		); err != nil {
			return nil, fmt.Errorf("movement repository: find all: %w", err)
		}
		movements = append(movements, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("movement repository: find all: %w", err)
	}
	return movements, nil
}
