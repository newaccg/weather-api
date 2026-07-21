package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"

	errs "github.com/newaccg/weather-api/internal/errors"
)

type middleware struct{
	rdb *redis.Client
	limiter *redis_rate.Limiter
	timeout time.Duration
	rateLimitPerMinute int
}

func NewMiddleware(client *redis.Client, timeout time.Duration, rateLimitPerMinute int) *middleware {
	return &middleware{
		rdb: client,
		limiter: redis_rate.NewLimiter(client),
		timeout: timeout,
		rateLimitPerMinute: rateLimitPerMinute,
	}
}

func (m *middleware) WithTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), m.timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rKey := fmt.Sprintf("%s:%s", r.RemoteAddr, r.URL.Path)
		ctx := r.Context()

		res, err := m.limiter.Allow(ctx, rKey, redis_rate.PerMinute(m.rateLimitPerMinute))
		if err != nil {
			errs.WriteError(w, errs.NewError(
				fmt.Errorf("could not get result from limiter: %w", err),
				"internal server error",
				http.StatusInternalServerError,
			))
			return
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", res.Limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", res.Remaining))

		if res.Allowed == 0 {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(res.RetryAfter.Seconds())))

			errs.WriteError(w, errs.NewError(
				errors.New("client reached request limit"),
				"too many requests",
				http.StatusTooManyRequests,
			))

			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
