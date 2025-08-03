package db

import (
	"context"
	"fafnir/auth-service/internal/config"
	"fafnir/auth-service/internal/db/generated"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

func NewDBConnection(cfg *config.Config) (*Database, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DB.URL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := generated.New(pool)

	return &Database{
		pool:    pool,
		queries: queries,
	}, nil
}

func (db *Database) GetQueries() *generated.Queries {
	return db.queries
}

func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *Database) GetPool() *pgxpool.Pool {
	return db.pool
}
