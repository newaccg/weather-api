package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/newaccg/weather-api/internal/config"
	errs "github.com/newaccg/weather-api/internal/errors"
)

type Service interface{
	GetWeatherByCity(context.Context, string) ([]byte, *errs.Error)
	GetHealth(ctx context.Context) []string
}

type middleware interface{
	WithTimeout(http.Handler, time.Duration) http.Handler
	WithRateLimit(http.Handler, int) http.Handler
}


type handler struct{
	service Service
	midware middleware
	timeouts config.Timeouts
	rateLimits config.RateLimitsPerMinute
}

func NewHandler(service Service, midware middleware, timeouts config.Timeouts, rateLimits config.RateLimitsPerMinute) *handler {
	return &handler{
		service: service,
		midware: midware,
		timeouts: timeouts,
		rateLimits: rateLimits,
	}
}

func (h *handler) RegisterRoutes() {
	var hdr http.Handler = http.HandlerFunc(h.GetWeather)
	hdr = h.midware.WithRateLimit(hdr, h.rateLimits.Weather)
	hdr = h.midware.WithTimeout(hdr, h.timeouts.Weather.Duration)
	http.Handle("GET /weather/", hdr)

	hdr = http.HandlerFunc(h.GetHealth)
	hdr = h.midware.WithTimeout(hdr, h.timeouts.Health.Duration)
	http.Handle("GET /health", hdr)
}

func (h *handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	unhealthy := h.service.GetHealth(r.Context())

	var msg struct {
		Status string `json:"message"`
		Code int `json:"code"`
		Unhealthy []string `json:"unhealthy,omitempty"`
	}

	w.Header().Set("Content-Type", "application/json")
	if len(unhealthy) != 0 {
		msg.Status = "unhealthy"
		msg.Code = http.StatusServiceUnavailable
		msg.Unhealthy = unhealthy

		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		msg.Status = "OK"
		msg.Code = http.StatusOK

		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(msg)
}

func (h *handler) GetWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.TrimPrefix(r.URL.Path, "/weather/")

	var err *errs.Error
	var weather []byte

	if city == "" {
		err = errs.NewError(
			errors.New("got empty city name. Returning 400"),
			"got empty city name",
			http.StatusBadRequest,
		)
	} else {
		log.Printf("got city: %s", city)
		weather, err = h.service.GetWeatherByCity(r.Context(), city)
	}

	if err != nil {
		errs.WriteError(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(weather)
}
