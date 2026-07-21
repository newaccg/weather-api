package handler

import (
	"context"
	"log"
	"net/http"
	"strings"
	"errors"

	errs "github.com/newaccg/weather-api/internal/errors"
)

type Service interface{
	GetWeatherByCity(context.Context, string) ([]byte, *errs.Error)
}

type middleware interface{
	WithTimeout(http.Handler) http.Handler
	RateLimit(http.Handler) http.Handler
}


type handler struct{
	service Service
	midware middleware
}

func NewHandler(service Service, midware middleware) *handler {
	return &handler{
		service: service,
		midware: midware,
	}
}

func (h *handler) RegisterRoutes() {
	var hdr http.Handler = http.HandlerFunc(h.GetWeather)
	hdr = h.midware.RateLimit(hdr)
	hdr = h.midware.WithTimeout(hdr)
	http.Handle("GET /weather/", hdr)
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
