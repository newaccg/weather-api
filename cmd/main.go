package main

import (
	"log"
	"net/http"

	"github.com/newaccg/weather-api/internal/handler"
	"github.com/newaccg/weather-api/internal/service"
)

func main() {
	var repo any = nil
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	h.RegisterRoutes()
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
