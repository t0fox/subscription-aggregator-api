package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/t0fox/subscription-aggregator-api/internal/config"
)

type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabase создаёт пул и ждёт готовности БД. Возвращает error, а не роняет процесс.
func NewDatabase(ctx context.Context, cfg *config.Config) (*Database, error) {
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

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Ждём, пока Postgres поднимется.
	var lastErr error
	for attempt := 1; attempt <= 30; attempt++ {
		if err := pool.Ping(ctx); err == nil {
			log.Println("Database connection established successfully")
			return &Database{Pool: pool}, nil
		} else {
			lastErr = err
			log.Printf("Database not ready, retry (%d/30): %v", attempt, err)
			time.Sleep(time.Second)
		}
	}

	pool.Close()
	return nil, fmt.Errorf("database not reachable after retries: %w", lastErr)
}

func (db *Database) Close() {
	db.Pool.Close()
	log.Println("Database connection closed")
}
