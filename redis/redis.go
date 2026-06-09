package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const defaultTimeout = 5 * time.Second

// RedisClient handles interactions with Redis
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new RedisClient instance
func NewRedisClient(host, port, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %v", err)
	}

	return &RedisClient{client: client}, nil
}

// IncrementButtonCount increments the count for a specific button and returns the new count
func (r *RedisClient) IncrementButtonCount(buttonCode string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	count, err := r.client.Incr(ctx, buttonCode).Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetButtonCount gets the count for a specific button
func (r *RedisClient) GetButtonCount(buttonCode string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	count, err := r.client.Get(ctx, buttonCode).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// ResetButtonCount resets the count for a specific button to 0
func (r *RedisClient) ResetButtonCount(buttonCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return r.client.Set(ctx, buttonCode, 0, 0).Err()
}

// Close closes the Redis client connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}
