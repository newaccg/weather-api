package handler
import (
	"net/http"
	"fmt"
)

type Service interface{
	GetWeather() string
}

type handler struct{
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) RegisterRoutes() {
	http.HandleFunc("GET /weather/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, h.service.GetWeather())
	})
}
