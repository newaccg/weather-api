package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"errors"
	"context"
	"log/slog"

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
		slog.Error(
			"could not get weather from cache",
			"error", err,
		)
	} else if result != nil {
		slog.Info("using cache")
		return result, nil
	}

	// if result is nil, fetch from third-party API and set to cache
	slog.Info("using third-party API")
	result, appError := s.fetchDataByCity(ctx, city, "GET")
	if appError != nil {
		return nil, appError
	}

	err = s.repo.SetWeatherByCity(ctx, city, result)
	if err != nil {
		slog.Error(
			"could not set weather to cache",
			"error", err,
		)
	}

	return result, nil
}

func (s *Service) GetHealth(ctx context.Context) []string {
	var unhealthy []string

	err := s.repo.GetCacheHealth(ctx)
	if err != nil {
		slog.Error(
			"cache check health failed",
			"error", err,
		)

		unhealthy = append(unhealthy, "cache")
	}

	// send HEAD - it returns light response
	_, appError := s.fetchDataByCity(ctx, "London", "HEAD")
	if appError != nil {
		slog.Error(
			"third-party API check health failed",
			"error", appError.InternalError,
		)
		unhealthy = append(unhealthy, "third-party API")
	}

	return unhealthy
}

func (s *Service) fetchDataByCity(ctx context.Context, city string, method string) ([]byte, *errs.Error) {
	parsed, err := url.Parse(s.apiUrl)
	if err != nil {
		return nil, errs.NewError(
			err,
			"could not get valid API URL. Please write the correct URL in the configuration file",
			"Could not get valid URL to third-party API",
			http.StatusInternalServerError,
		)
	}

	parsed = parsed.JoinPath("/", city)
	path := parsed.String()

	path = fmt.Sprintf("%s?key=%s", path, s.apiKey)

	slog.Info(
		"outcoming request",
		"method", method,
		"URL", path,
	)

	request, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return nil, errs.InternalServerError(
			err,
			"could not form request",
		)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		// if we got timeout...
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, errs.NewError(
				err,
				"third-party API not responding",
				"third-party API not responding",
				http.StatusGatewayTimeout,
			)
		}

		// if not a timeout, it's a different error, return 502
		return nil, errs.NewError(
			err,
			"could not send request to third-party API",
			"could not send request to third-party API",
			http.StatusBadGateway,
		)
	}
	defer resp.Body.Close()

	slog.Info(
		"got response",
		"HTTPCode", resp.StatusCode,
	)

	if resp.StatusCode != 200 {
		appError := errs.NewError(
			fmt.Errorf("third-party API returned non successful code %d", resp.StatusCode),
			"third-party API returned non successful code",
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
			err,
			"could not read body of API response",
			"could not read API response",
			http.StatusInternalServerError,
		)
	}

	return result, nil
}
