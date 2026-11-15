package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"odontoestoque/infra/database/queries"
	"odontoestoque/internal/models"
)

type LanguageRepository struct {
	db *pgxpool.Pool
}

func NewLanguageRepository(db *pgxpool.Pool) *LanguageRepository {
	return &LanguageRepository{db: db}
}

// FindAll executa queries/idiomas/select_all.sql.
func (r *LanguageRepository) FindAll(ctx context.Context) ([]models.Language, error) {
	rows, err := r.db.Query(ctx, queries.IdiomaSelectAll)
	if err != nil {
		return nil, fmt.Errorf("language repository: find all: %w", err)
	}
	defer rows.Close()

	languages := []models.Language{}
	for rows.Next() {
		var l models.Language
		if err := rows.Scan(&l.Codigo, &l.Nome, &l.Regiao, &l.Ativo, &l.CriadoEm); err != nil {
			return nil, fmt.Errorf("language repository: find all: %w", err)
		}
		languages = append(languages, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("language repository: find all: %w", err)
	}
	return languages, nil
}
