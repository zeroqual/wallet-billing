package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgres(ctx context.Context, dsn string) (*pgxpool.Pool, error) {

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	ctxTime, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := pool.Ping(ctxTime); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping pool: %w", err)
	}

	return pool, nil
}
