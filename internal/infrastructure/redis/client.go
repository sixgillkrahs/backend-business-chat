package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisConnection(addr string, password string, db int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        password,
		DB:              db,
		PoolSize:        50,              // Số lượng kết nối tối đa trong pool
		MinIdleConns:    10,              // Số kết nối rảnh luôn duy trì
		DialTimeout:     5 * time.Second, // Thời gian chờ tối đa khi kết nối
		ReadTimeout:     3 * time.Second, // Thời gian chờ đọc dữ liệu
		WriteTimeout:    3 * time.Second, // Thời gian chờ ghi dữ liệu
		ConnMaxIdleTime: 5 * time.Minute, // Đóng kết nối rảnh sau 5 phút
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}
	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
