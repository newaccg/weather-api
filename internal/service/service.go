package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"errors"
	"context"
	"log"

	errs "github.com/newaccg/weather-api/internal/errors"
)

type Repository interface{
	GetWeatherByCity(context.Context, string) ([]byte, error)
	SetWeatherByCity(context.Context, string, []byte) error
	GetCacheHealth(context.Context) error
}

type Service struct{
	apiUrl string
	apiKey string
	repo Repository
}

func NewService(repo Repository, apiUrl, apiKey string) *Service {
	return &Service{
		repo: repo,
		apiUrl: apiUrl,
		apiKey: apiKey,
	}
}

func (s *Service) GetWeatherByCity(ctx context.Context, city string) ([]byte, *errs.Error) {
	result, err := s.repo.GetWeatherByCity(ctx, city)

	if err != nil {
		log.Printf("[ERR] could not get weather from cache: %s", err)
	} else if result != nil {
		log.Println("using cache")
		return result, nil
	}

	// if result is nil, fetch from third-party API and set to cache
	log.Println("using third-party API")
	result, appError := s.fetchDataByCity(ctx, city, "GET")
	if appError != nil {
		return nil, appError
	}

	err = s.repo.SetWeatherByCity(ctx, city, result)
	if err != nil {
		log.Printf("[ERR] could not set weather to cache: %s", err)
	}

	return result, nil
}

func (s *Service) GetHealth(ctx context.Context) []string {
	var unhealthy []string

	err := s.repo.GetCacheHealth(ctx)
	if err != nil {
		log.Printf("[ERR] cache check health failed: %s", err)
		unhealthy = append(unhealthy, "cache")
	}

	// send HEAD - it returns light response
	_, appError := s.fetchDataByCity(ctx, "London", "HEAD")
	if appError != nil {
		log.Printf("[ERR] third-party API check health failed: %s", appError.InternalError)
		unhealthy = append(unhealthy, "third-party API")
	}

	return unhealthy
}

func (s *Service) fetchDataByCity(ctx context.Context, city string, method string) ([]byte, *errs.Error) {
	parsed, err := url.Parse(s.apiUrl)
	if err != nil {
		return nil, errs.NewError(
			fmt.Errorf("could not get valid API URL: %w. Please write the correct URL in the configuration file", err),
			"Could not get valid URL to third-party API",
			http.StatusInternalServerError,
		)
	}

	parsed = parsed.JoinPath("/", city)
	path := parsed.String()

	path = fmt.Sprintf("%s?key=%s", path, s.apiKey)

	log.Printf("sending %s to %s", method, path)

	request, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return nil, errs.NewError(
			fmt.Errorf("could not form request: %w", err),
			"internal server error",
			http.StatusInternalServerError,
		)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		// if we got timeout...
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
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

	log.Println("got response")

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
