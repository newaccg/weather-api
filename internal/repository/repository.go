package repository

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/newaccg/weather-api/internal/config"
	"github.com/redis/go-redis/v9"
)

type repository struct {
	expirationTime time.Duration
	rdb            *redis.Client
	limiter        redis_rate.Limiter
}

func NewRepository(cfg *config.DBConfig) *repository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DatabaseNumber,
	})

	return &repository{
		rdb:            rdb,
		expirationTime: cfg.KeyExpirationTime.Duration,
		limiter:        *redis_rate.NewLimiter(rdb),
	}
}

func (r *repository) GetWeatherByCity(ctx context.Context, city string) ([]byte, error) {
	val, err := r.rdb.Get(ctx, city).Bytes()
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
	err := r.rdb.Set(ctx, city, data, r.expirationTime).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetCacheHealth(ctx context.Context) error {
	return r.rdb.Ping(ctx).Err()
}

func (r *repository) AllowRequest(ctx context.Context, key string, limit int) (*redis_rate.Result, error) {
	return r.limiter.Allow(ctx, key, redis_rate.PerMinute(limit))
}
