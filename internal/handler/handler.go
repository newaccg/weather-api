package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	errs "github.com/newaccg/weather-api/internal/errors"
)

type Service interface{
	GetWeatherByCity(context.Context, string) ([]byte, *errs.Error)
}

type handler struct{
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) RegisterRoutes() {
	http.HandleFunc("GET /weather/", withTimeout(h.GetWeather, 5 * time.Second))
}

func withTimeout(handle http.HandlerFunc, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		handle(w, r.WithContext(ctx))
	}
}

func (h *handler) GetWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.TrimPrefix(r.URL.Path, "/weather/")
	log.Printf("got city: %s", city)

	w.Header().Set("Content-Type", "application/json")
	weather, err := h.service.GetWeatherByCity(r.Context(), city)
	if err != nil {
		log.Printf("[ERR] %s", err.InternalError)

		var js struct {
			Error string `json:"error"`
			Code int `json:"code"`
		}

		js.Error = err.ErrorMessage
		js.Code = err.HttpCode

		w.WriteHeader(err.HttpCode)
		json.NewEncoder(w).Encode(js)
		return
	}

	w.Write(weather)
}
