package main

import (
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"

	"github.com/newaccg/weather-api/internal/config"
	"github.com/newaccg/weather-api/internal/handler"
	"github.com/newaccg/weather-api/internal/middleware"
	"github.com/newaccg/weather-api/internal/repository"
	"github.com/newaccg/weather-api/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
		Password: "",
		DB: 0,
	})

	repo := repository.NewRepository(rdb, cfg.ExpirationTime.Duration)
	svc := service.NewService(repo, cfg.ApiUrl, cfg.ApiKey)
	midware := middleware.NewMiddleware(rdb)
	h := handler.NewHandler(svc, midware, cfg.Timeouts, cfg.RateLimits)

	h.RegisterRoutes()
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, nil))
}
