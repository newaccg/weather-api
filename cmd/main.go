package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/newaccg/weather-api/internal/client"
	"github.com/newaccg/weather-api/internal/config"
	"github.com/newaccg/weather-api/internal/handler"
	"github.com/newaccg/weather-api/internal/middleware"
	"github.com/newaccg/weather-api/internal/repository"
	"github.com/newaccg/weather-api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
		Password: "",
		DB: 0,
	})

	httpClient := client.NewMaskClient(cfg.ApiKey, &http.Client{})

	repo := repository.NewRepository(rdb, cfg.ExpirationTime.Duration)
	svc := service.NewService(repo, cfg.ApiUrl, cfg.ApiKey, httpClient)
	midware := middleware.NewMiddleware(rdb)
	h := handler.NewHandler(svc, midware, cfg.Timeouts, cfg.RateLimits)

	mux := h.RegisterRoutes()
	server := &http.Server{
		Addr: cfg.ServerAddress,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
