package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"log"

	errs "github.com/newaccg/weather-api/internal/errors"
)

type Service interface{
	GetWeatherByCity(string) ([]byte, *errs.Error)
}

type handler struct{
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) RegisterRoutes() {
	http.HandleFunc("GET /weather/", h.GetWeather)
}

func (h *handler) GetWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.TrimPrefix(r.URL.Path, "/weather/")
	log.Printf("got city: %s", city)

	w.Header().Set("Content-Type", "application/json")
	weather, err := h.service.GetWeatherByCity(city)
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
