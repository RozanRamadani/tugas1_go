package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/config"
)

// NewPool membuat connection pool PostgreSQL
// sekaligus memastikan database dapat dihubungi.
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {

	host := config.GetEnv("DB_HOST", "localhost")
	port := config.GetEnv("DB_PORT", "5432")
	user := config.GetEnv("DB_USER", "postgres")
	password := config.GetEnv("DB_PASSWORD", "")
	dbname := config.GetEnv("DB_NAME", "praktikum_backend")
	sslmode := config.GetEnv("DB_SSLMODE", "disable")

	maxConns := config.GetEnv("DB_MAX_CONNS", "10")
	minConns := config.GetEnv("DB_MIN_CONNS", "2")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		password,
		host,
		port,
		dbname,
		sslmode,
	)

	cfg, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	// Konfigurasi jumlah koneksi pool.
	fmt.Sscanf(maxConns, "%d", &cfg.MaxConns)
	fmt.Sscanf(minConns, "%d", &cfg.MinConns)

	// Timeout koneksi.
	cfg.MaxConnLifetime = 1 * time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	// Pastikan PostgreSQL benar-benar dapat dihubungi.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return pool, nil
}
