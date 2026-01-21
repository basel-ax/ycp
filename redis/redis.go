package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

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

	// Test the connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %v", err)
	}

	return &RedisClient{client: client}, nil
}

// IncrementButtonCount increments the count for a specific button
func (r *RedisClient) IncrementButtonCount(buttonCode string) error {
	ctx := context.Background()
	_, err := r.client.Incr(ctx, buttonCode).Result()
	return err
}

// GetButtonCount gets the count for a specific button
func (r *RedisClient) GetButtonCount(buttonCode string) (int, error) {
	ctx := context.Background()
	count, err := r.client.Get(ctx, buttonCode).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// IncrementTotalCommands increments the total number of transmitted commands
func (r *RedisClient) IncrementTotalCommands() error {
	ctx := context.Background()
	_, err := r.client.Incr(ctx, "total_commands").Result()
	return err
}

// GetTotalCommands gets the total number of transmitted commands
func (r *RedisClient) GetTotalCommands() (int, error) {
	ctx := context.Background()
	count, err := r.client.Get(ctx, "total_commands").Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// CheckLimitReached checks if the total limit on transmitted commands has been reached
func (r *RedisClient) CheckLimitReached(limit int) (bool, error) {
	count, err := r.GetTotalCommands()
	if err != nil {
		return false, err
	}
	return count >= limit, nil
}

// Close closes the Redis client connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}
