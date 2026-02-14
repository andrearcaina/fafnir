package db

import (
	"context"
	"fafnir/shared/pkg/logger"
	"fafnir/user-service/internal/config"
	"fafnir/user-service/internal/db/generated"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
	logger  *logger.Logger
}

func New(cfg *config.Config, logger *logger.Logger) (*Database, error) {
	var pool *pgxpool.Pool
	var err error

	const maxRetries = 10
	const retryInterval = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		pool, err = pgxpool.New(context.Background(), cfg.DB.URL)
		if err != nil {
			logger.Error(context.Background(), "Failed to create connection pool", "attempt", attempt, "max_retries", maxRetries, "error", err)
			if attempt < maxRetries {
				time.Sleep(retryInterval)
				continue
			}
			return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = pool.Ping(ctx)
		cancel()

		if err == nil {
			logger.Info(context.Background(), "Successfully connected to database", "attempt", attempt)
			queries := generated.New(pool)
			return &Database{
				pool:    pool,
				queries: queries,
				logger:  logger,
			}, nil
		}

		logger.Error(context.Background(), "Failed to ping database", "attempt", attempt, "max_retries", maxRetries, "error", err)

		pool.Close()

		if attempt < maxRetries {
			logger.Info(context.Background(), "Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
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
