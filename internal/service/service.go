package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	errs "github.com/newaccg/weather-api/internal/errors"
	"github.com/newaccg/weather-api/internal/model"
)

type Repository interface{
	GetWeatherByCity(context.Context, string) ([]byte, error)
	SetWeatherByCity(context.Context, string, []byte) error
	GetCacheHealth(context.Context) error
}

type Client interface{
	Do(*http.Request) (*http.Response, error)
}

type Service struct{
	apiUrl string
	apiKey string
	repo Repository
	client Client
}

func NewService(repo Repository, apiUrl, apiKey string, client Client) *Service {
	return &Service{
		repo: repo,
		apiUrl: apiUrl,
		apiKey: apiKey,
		client: client,
	}
}

func (s *Service) GetWeatherByCity(ctx context.Context, city string) (*model.WeatherResponse, *errs.Error) {
	data, err := s.repo.GetWeatherByCity(ctx, city)

	if err != nil {
		slog.Error(
			"could not get weather from cache",
			"error", err,
		)
	} else if data != nil {
		slog.Info("using cache")
		return encodeResponse(data)
	}

	// if result is nil, fetch from third-party API and set to cache
	slog.Info("using third-party API")
	data, appError := s.fetchDataByCity(ctx, city, "GET")
	if appError != nil {
		return nil, appError
	}

	err = s.repo.SetWeatherByCity(ctx, city, data)
	if err != nil {
		slog.Error(
			"could not set weather to cache",
			"error", err,
		)
	}

	return encodeResponse(data)
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

func encodeResponse(data []byte) (*model.WeatherResponse, *errs.Error) {
	var result *model.WeatherResponse
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, errs.NewError(
			err,
			"could not marshal data to response",
			"got invalid API response",
			http.StatusBadGateway,
		)
	}

	return result, nil
}

func (s *Service) fetchDataByCity(ctx context.Context, city string, method string) ([]byte, *errs.Error) {
	// forming URL

	parsed, err := url.Parse(s.apiUrl)
	if err != nil {
		return nil, errs.NewError(
			err,
			"could not get valid API URL. Please write the correct URL in the configuration file",
			"Could not get valid URL to third-party API",
			http.StatusInternalServerError,
		)
	}

	parsed = parsed.JoinPath(city)

	path := parsed.String()

	// logging URL without API key
	slog.Info(
		"outcoming request",
		"method", method,
		"URL", path,
	)

	// now, adding API key to URL
	q := parsed.Query()
	q.Set("key", s.apiKey)

	// writing encoded to URL
	parsed.RawQuery = q.Encode()

	path = parsed.String()

	request, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return nil, errs.InternalServerError(
			err,
			"could not form request",
		)
	}

	resp, err := s.client.Do(request)
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
			appError.OutputMessage = "city not found or invalid or invalid API format"
			appError.HttpCode = http.StatusBadRequest

		case 401:
			appError.OutputMessage = "your API key is invalid"
			appError.HttpCode = http.StatusUnauthorized

		case 404:
			appError.OutputMessage = "invalid API format"
			appError.HttpCode = http.StatusBadRequest

		case 429:
			appError.OutputMessage = "your account has exceeded the set limits"
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
