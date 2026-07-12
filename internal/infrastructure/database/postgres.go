package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

func NewPostgresConnection(ctx context.Context, connStr string) (*PostgresDB, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}
	config.MaxConns = 25                      // Số lượng kết nối tối đa
	config.MinConns = 5                       // Số kết nối rảnh luôn duy trì
	config.MaxConnLifetime = 30 * time.Minute // Thời gian sống tối đa của 1 kết nối
	config.MaxConnIdleTime = 5 * time.Minute  // Thời gian rảnh tối đa trước khi tự đóng
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	return &PostgresDB{Pool: pool}, nil
}

func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
