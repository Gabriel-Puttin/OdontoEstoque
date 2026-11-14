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

// Códigos de erro do PostgreSQL usados aqui.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create executa queries/usuarios/insert.sql e devolve o usuário criado.
func (r *UserRepository) Create(ctx context.Context, nome, email, senhaHash string, idioma *string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, queries.UsuarioInsert, nome, email, senhaHash, idioma).Scan(
		&u.ID, &u.Nome, &u.Email, &u.IdiomaCodigo, &u.CriadoEm, &u.AtualizadoEm,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == "uq_usuarios_email":
				return nil, models.ErrEmailAlreadyExists
			case pgErr.Code == pgForeignKeyViolation:
				return nil, models.ErrInvalidLanguage
			}
		}
		return nil, fmt.Errorf("user repository: create: %w", err)
	}
	return &u, nil
}

// FindByEmail executa queries/usuarios/select_by_email.sql.
// Retorna models.ErrUserNotFound quando não há linha.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, queries.UsuarioSelectByEmail, email).Scan(
		&u.ID, &u.Nome, &u.Email, &u.SenhaHash, &u.IdiomaCodigo, &u.CriadoEm, &u.AtualizadoEm,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user repository: find by email: %w", err)
	}
	return &u, nil
}
