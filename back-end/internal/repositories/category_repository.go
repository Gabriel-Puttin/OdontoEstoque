package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"odontoestoque/infra/database/queries"
	"odontoestoque/internal/models"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// FindAll executa queries/categorias/select_all.sql.
func (r *CategoryRepository) FindAll(ctx context.Context) ([]models.Category, error) {
	rows, err := r.db.Query(ctx, queries.CategoriaSelectAll)
	if err != nil {
		return nil, fmt.Errorf("category repository: find all: %w", err)
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var c models.Category
		c.Itens = new(int64)
		if err := rows.Scan(&c.ID, &c.Nome, c.Itens, &c.CriadoEm, &c.AtualizadoEm); err != nil {
			return nil, fmt.Errorf("category repository: find all: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category repository: find all: %w", err)
	}
	return categories, nil
}

// FindByID executa queries/categorias/select_by_id.sql.
func (r *CategoryRepository) FindByID(ctx context.Context, id int16) (*models.Category, error) {
	c, err := scanCategory(r.db.QueryRow(ctx, queries.CategoriaSelectByID, id))
	if err != nil {
		return nil, fmt.Errorf("category repository: find by id: %w", err)
	}
	return c, nil
}

// FindProducts executa queries/categorias/select_produtos.sql.
func (r *CategoryRepository) FindProducts(ctx context.Context, categoryID int16) ([]models.CategoryProduct, error) {
	rows, err := r.db.Query(ctx, queries.CategoriaSelectProdutos, categoryID)
	if err != nil {
		return nil, fmt.Errorf("category repository: find products: %w", err)
	}
	defer rows.Close()

	products := []models.CategoryProduct{}
	for rows.Next() {
		var p models.CategoryProduct
		if err := rows.Scan(
			&p.ID, &p.Nome, &p.Variacao, &p.NomeExibicao, &p.TipoProdutoCodigo, &p.TipoProduto,
			&p.Grupo, &p.UnidadeMedida, &p.QuantidadeAtual, &p.EstoqueMinimo, &p.Validade,
		); err != nil {
			return nil, fmt.Errorf("category repository: find products: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category repository: find products: %w", err)
	}
	return products, nil
}

// Create executa queries/categorias/insert.sql.
func (r *CategoryRepository) Create(ctx context.Context, nome string) (*models.Category, error) {
	c, err := scanCategory(r.db.QueryRow(ctx, queries.CategoriaInsert, nome))
	if err != nil {
		return nil, fmt.Errorf("category repository: create: %w", err)
	}
	return c, nil
}

// Update executa queries/categorias/update.sql.
func (r *CategoryRepository) Update(ctx context.Context, id int16, nome string) (*models.Category, error) {
	c, err := scanCategory(r.db.QueryRow(ctx, queries.CategoriaUpdate, id, nome))
	if err != nil {
		return nil, fmt.Errorf("category repository: update: %w", err)
	}
	return c, nil
}

// Delete executa queries/categorias/delete.sql (exclusão lógica).
// Sem linha -> ErrCategoryNotFound; com produtos ativos -> ErrCategoryHasProducts.
func (r *CategoryRepository) Delete(ctx context.Context, id int16) error {
	var temProdutos bool
	err := r.db.QueryRow(ctx, queries.CategoriaDelete, id).Scan(&temProdutos)
	if err != nil {
		return fmt.Errorf("category repository: delete: %w", translateCategoryErr(err))
	}
	if temProdutos {
		return models.ErrCategoryHasProducts
	}
	return nil
}

// scanCategory lê id, nome, criado_em, atualizado_em e traduz os erros do Postgres.
func scanCategory(row pgx.Row) (*models.Category, error) {
	var c models.Category
	if err := row.Scan(&c.ID, &c.Nome, &c.CriadoEm, &c.AtualizadoEm); err != nil {
		return nil, translateCategoryErr(err)
	}
	return &c, nil
}

func translateCategoryErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ErrCategoryNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == "uq_categorias_nome" {
		return models.ErrCategoryNameExists
	}
	return err
}
