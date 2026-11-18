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

const pgCheckViolation = "23514"

// Mapeia as FKs (nomes padrão do Postgres) para o campo do JSON.
var productFKFields = map[string]string{
	"produtos_tipo_produto_codigo_fkey":     "tipo_produto_codigo",
	"produtos_categoria_id_fkey":            "categoria_id",
	"produtos_unidade_medida_id_fkey":       "unidade_medida_id",
	"movimentacoes_estoque_usuario_id_fkey": "usuario_id",
}

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

// FindAll executa queries/produtos/select_all.sql.
func (r *ProductRepository) FindAll(ctx context.Context) ([]models.Product, error) {
	rows, err := r.db.Query(ctx, queries.ProdutoSelectAll)
	if err != nil {
		return nil, fmt.Errorf("product repository: find all: %w", err)
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("product repository: find all: %w", err)
		}
		products = append(products, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("product repository: find all: %w", err)
	}
	return products, nil
}

// FindByID executa queries/produtos/select_by_id.sql.
func (r *ProductRepository) FindByID(ctx context.Context, id int64) (*models.Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx, queries.ProdutoSelectByID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("product repository: find by id: %w", err)
	}
	return p, nil
}

// Create, numa transação:
//  1. produtos/insert.sql (saldo 0)
//  2. movimentacoes/insert.sql como 'entrada' 0 -> quantidade, se quantidade > 0
func (r *ProductRepository) Create(ctx context.Context, in models.ProductInput, quantidade int32, usuarioID *int64) (int64, error) {
	var id int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, queries.ProdutoInsert,
			in.TipoProdutoCodigo, in.CategoriaID, in.UnidadeMedidaID, in.Nome, in.Variacao, in.EstoqueMinimo, in.Validade,
		).Scan(&id); err != nil {
			return err
		}
		if quantidade > 0 {
			if _, err := tx.Exec(ctx, queries.MovimentacaoInsert, id, usuarioID, "entrada", 0, quantidade); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("product repository: create: %w", translateProductErr(err))
	}
	return id, nil
}

// Update, numa transação:
//  1. produtos/select_saldo_for_update.sql (trava a linha e lê o saldo)
//  2. produtos/update.sql (dados cadastrais)
//  3. movimentacoes/insert.sql saldo -> novaQuantidade, se informada e diferente:
//     'entrada' quando o saldo sobe, 'saida' quando desce
func (r *ProductRepository) Update(ctx context.Context, id int64, in models.ProductInput, novaQuantidade *int32, usuarioID *int64) error {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var saldo int32
		if err := tx.QueryRow(ctx, queries.ProdutoSelectSaldoForUpdate, id).Scan(&saldo); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, queries.ProdutoUpdate,
			id, in.TipoProdutoCodigo, in.CategoriaID, in.UnidadeMedidaID, in.Nome, in.Variacao, in.EstoqueMinimo, in.Validade,
		); err != nil {
			return err
		}
		if novaQuantidade != nil && *novaQuantidade != saldo {
			tipo := "entrada"
			if *novaQuantidade < saldo {
				tipo = "saida"
			}
			if _, err := tx.Exec(ctx, queries.MovimentacaoInsert, id, usuarioID, tipo, saldo, *novaQuantidade); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("product repository: update: %w", translateProductErr(err))
	}
	return nil
}

// Delete executa queries/produtos/delete.sql (exclusão lógica).
func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, queries.ProdutoDelete, id)
	if err != nil {
		return fmt.Errorf("product repository: delete: %w", translateProductErr(err))
	}
	if tag.RowsAffected() == 0 {
		return models.ErrProductNotFound
	}
	return nil
}

func scanProduct(row pgx.Row) (*models.Product, error) {
	var p models.Product
	err := row.Scan(
		&p.ID, &p.Nome, &p.Variacao, &p.NomeExibicao, &p.TipoProdutoCodigo, &p.TipoProduto, &p.Grupo,
		&p.CategoriaID, &p.Categoria, &p.UnidadeMedidaID, &p.UnidadeMedida,
		&p.QuantidadeAtual, &p.EstoqueMinimo, &p.Situacao, &p.Faltam,
		&p.Validade, &p.DiasParaVencer, &p.CriadoEm, &p.AtualizadoEm,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func translateProductErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ErrProductNotFound
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case pgUniqueViolation:
		if pgErr.ConstraintName == "uq_produtos_nome_variacao" {
			return models.ErrProductExists
		}
	case pgForeignKeyViolation:
		if field, ok := productFKFields[pgErr.ConstraintName]; ok {
			return &models.ValidationError{Fields: map[string]string{field: "não existe"}}
		}
	case pgCheckViolation:
		// RAISE EXCEPTION dos triggers não tem ConstraintName; a mensagem já é legível.
		if pgErr.ConstraintName == "" {
			return &models.RuleError{Message: pgErr.Message}
		}
	}
	return err
}
