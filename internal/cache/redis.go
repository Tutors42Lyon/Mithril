package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient(url string, password string, db int) (*RedisClient, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	opts.Password = password
	opts.DB = db
	opts.PoolSize = 10

	client := redis.NewClient(opts)
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{client: client, ctx: ctx}, nil
}

func (rc *RedisClient) Get(key string) (string, error) {
	val, err := rc.client.Get(rc.ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	return val, err
}

func (rc *RedisClient) Set(key string, value string, ttl time.Duration) error {
	return rc.client.Set(rc.ctx, key, value, ttl).Err()
}

func (rc *RedisClient) Delete(key string) error {
	return rc.client.Del(rc.ctx, key).Err()
}

func (rc *RedisClient) Exists(key string) (bool, error) {
	count, err := rc.client.Exists(rc.ctx, key).Result()
	return count > 0, err
}

func (rc *RedisClient) GetList(key string) ([]string, error) {
	return rc.client.LRange(rc.ctx, key, 0, -1).Result()
}

func (rc *RedisClient) AddToList(key string, values ...string) error {
	// Convert []string to []interface{} for RPush
	interfaceSlice := make([]interface{}, len(values))
	for i, v := range values {
		interfaceSlice[i] = v
	}
	return rc.client.RPush(rc.ctx, key, interfaceSlice...).Err()
}

