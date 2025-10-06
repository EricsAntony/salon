package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// NewConnection creates a new database connection pool
func NewConnection(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	config.MaxConns = 30
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Database connection established successfully")
	return pool, nil
}

// HealthCheck performs a health check on the database connection
func HealthCheck(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int
	err := pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	return nil
}
