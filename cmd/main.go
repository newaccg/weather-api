package main

import (
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/newaccg/weather-api/internal/config"
	"github.com/newaccg/weather-api/internal/handler"
	"github.com/newaccg/weather-api/internal/repository"
	"github.com/newaccg/weather-api/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
		Password: "",
		DB: 0,
	})

	repo := repository.NewRepository(redisClient, time.Hour * time.Duration(cfg.ExpirationTime))
	svc := service.NewService(repo, cfg.ApiUrl, cfg.ApiKey)
	h := handler.NewHandler(svc)

	h.RegisterRoutes()
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, nil))
}
