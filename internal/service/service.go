package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"net"
	"errors"
	"context"
	"log"

	errs "github.com/newaccg/weather-api/internal/errors"
)

type Repository interface{
}

type Service struct{
	ApiUrl string
	ApiKey string
	Repo Repository
}

func NewService(repo Repository, apiUrl, apiKey string) *Service {
	return &Service{
		Repo: repo,
		ApiUrl: apiUrl,
		ApiKey: apiKey,
	}
}

func (s *Service) GetWeatherByCity(city string) ([]byte, *errs.Error) {
	url, err := url.JoinPath(s.ApiUrl, "/", city)
	if err != nil {
		return nil, errs.NewError(
			fmt.Errorf("could not get valid API URL: %w", err),
			"bad URl of API or bad city name",
			http.StatusBadRequest,
		)
	}

	url = fmt.Sprintf("%s?key=%s", url, s.ApiKey)

	log.Printf("sending GET to %s", url)

	resp, err := http.Get(url)
	if err != nil {
		// if we got timeout...
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
			return nil, errs.NewError(
				fmt.Errorf("third-party API not responding: %w", err),
				"third-party API not responding",
				http.StatusGatewayTimeout,
			)
		}

		// if not a timeout, it's a different error, return 502
		return nil, errs.NewError(
			fmt.Errorf("could not send request to third-party API: %w", err),
				"could not send request to third-party API",
				http.StatusBadGateway,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		appError := errs.NewError(
			fmt.Errorf("the third-party API returned non successful code %d", resp.StatusCode),
			"",
			http.StatusBadGateway,
		)

		switch resp.StatusCode {
		case 400:
			appError.ErrorMessage = "city not found or invalid or invalid API format"
			appError.HttpCode = http.StatusBadRequest

		case 401:
			appError.ErrorMessage = "your API key is invalid"
			appError.HttpCode = http.StatusUnauthorized

		case 404:
			appError.ErrorMessage = "invalid API format"
			appError.HttpCode = http.StatusBadRequest

		case 429:
			appError.ErrorMessage = "your account has exceeded the set limits"
			appError.HttpCode = http.StatusTooManyRequests
		}

		return nil, appError
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errs.NewError(
			fmt.Errorf("could not read body of API response: %w", err),
			"could not read API response",
			http.StatusInternalServerError,
		)
	}

	return result, nil
}

