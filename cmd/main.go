package main

import (
	"log/slog"
	"net/http"
	"os"

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

	cfg, err := config.LoadConfig("internal/config/config.json")
	if err != nil {
		slog.Error(
			"could not load config",
			"error", err,
		)

		os.Exit(1)
	}

	httpClient := client.NewMaskClient(cfg.ApiKey, &http.Client{})

	repo := repository.NewRepository(&cfg.DB)
	svc := service.NewService(repo, cfg.ApiUrl, cfg.ApiKey, httpClient)
	midware := middleware.NewMiddleware(repo)
	h := handler.NewHandler(svc, midware, cfg.Timeouts, cfg.RateLimits)

	mux := h.RegisterRoutes()
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: mux,
	}

	slog.Error(
		"server error",
		"error", server.ListenAndServe(),
	)
}
