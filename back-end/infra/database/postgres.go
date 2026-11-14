package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect abre um pool de conexões com o PostgreSQL usando o pgx (driver puro, sem ORM).
func Connect(ctx context.Context, dsn string, logger *slog.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("database: falha ao configurar pool: %w", err)
	}

	// Testa a conexão imediatamente
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping falhou: %w", err)
	}

	logger.Info("PostgreSQL conectado")
	return pool, nil
}
