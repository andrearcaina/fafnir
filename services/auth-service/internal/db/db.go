package db

import (
	"context"
	"fafnir/auth-service/internal/config"
	"fafnir/auth-service/internal/db/generated"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

func New(cfg *config.Config) (*Database, error) {
	var pool *pgxpool.Pool
	var err error

	const maxRetries = 10
	const retryInterval = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		pool, err = pgxpool.New(context.Background(), cfg.DB.URL)
		if err != nil {
			log.Printf("Attempt %d/%d: Failed to create connection pool: %v", attempt, maxRetries, err)
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
			log.Printf("Successfully connected to database on attempt %d", attempt)
			queries := generated.New(pool)
			return &Database{
				pool:    pool,
				queries: queries,
			}, nil
		}

		log.Printf("Attempt %d/%d: Failed to ping database: %v", attempt, maxRetries, err)

		pool.Close()

		if attempt < maxRetries {
			log.Printf("Retrying in %v...", retryInterval)
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
