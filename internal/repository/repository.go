package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type repository struct {
	expirationTime time.Duration
	redisClient    *redis.Client
}

func NewRepository(client *redis.Client, expiration time.Duration) *repository {
	return &repository{
		redisClient:    client,
		expirationTime: expiration,
	}
}

func (r *repository) GetWeatherByCity(ctx context.Context, city string) ([]byte, error) {
	val, err := r.redisClient.Get(ctx, city).Bytes()
	if err != nil {
		if err == redis.Nil { // if the key doesn't exist...
			// there's no error and no data to return
			return nil, nil
		}

		return nil, err
	}

	return val, nil
}

func (r *repository) SetWeatherByCity(ctx context.Context, city string, data []byte) error {
	err := r.redisClient.Set(ctx, city, data, r.expirationTime).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetCacheHealth(ctx context.Context) error {
	return r.redisClient.Ping(ctx).Err()
}
