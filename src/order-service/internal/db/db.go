package db

import (
	"context"
	"fmt"
	"time"

	"fafnir/order-service/internal/config"
	"fafnir/order-service/internal/db/generated"
	"fafnir/shared/pkg/logger"

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

func (db *Database) ExecMultiTx(ctx context.Context, fn func(*generated.Queries) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	queries := generated.New(tx)
	if err := fn(queries); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %v; rollback failed: %w", err, rollbackErr)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
