package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/t0fox/subscription-aggregator-api/internal/config"
)

// Database wraps a pgx connection pool.
type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabase creates and verifies a connection pool.
// It returns an error instead of calling log.Fatal, so the package can be reused
// as a library and the caller decides how to handle failures.
func NewDatabase(cfg *config.Config) (*Database, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=prefer",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.HealthCheckPeriod = time.Minute
	poolConfig.MaxConnLifetime = 2 * time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Database{Pool: pool}, nil
}

// Close releases all pool connections.
func (db *Database) Close() {
	db.Pool.Close()
}
