package main

import (
	"log"
	"net/http"

	"github.com/newaccg/weather-api/internal/handler"
	"github.com/newaccg/weather-api/internal/service"
	"github.com/newaccg/weather-api/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	var repo any = nil
	svc := service.NewService(repo, cfg.ApiUrl, cfg.ApiKey)
	h := handler.NewHandler(svc)

	h.RegisterRoutes()
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, nil))
}
