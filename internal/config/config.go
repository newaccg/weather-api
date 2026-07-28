package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress  string `json:"serverAddress"`
	RedisAddress   string `json:"redisAddress"`
	ApiUrl         string `json:"visualCrossingApiUrl"`
	ApiKey         string
	ExpirationTime Duration `json:"expirationTime"`

	Timeouts   Timeouts            `json:"timeouts"`
	RateLimits RateLimitsPerMinute `json:"rateLimitsPerMinute"`
}

type Timeouts struct {
	Weather Duration `json:"weather"`
	Health  Duration `json:"health"`
}

type RateLimitsPerMinute struct {
	Weather int `json:"weather"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}

	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

func LoadConfig() (*Config, error) {
	var config *Config

	err := godotenv.Load()
	if err != nil {
		return nil,
			fmt.Errorf("could not load environment variables: %w. Please make sure that the .env file is located in the project root and mathes the .env.example file",
				err,
			)
	}

	key := os.Getenv("VISUAL_CROSSING_API_KEY")
	if key == "" {
		return nil, errors.New("Visual Crossing API key is not specified")
	}

	js, err := os.ReadFile("internal/config/config.json")
	if err != nil {
		return nil, fmt.Errorf("could not read config.json: %w", err)
	}

	err = json.Unmarshal(js, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config.json: %w", err)
	}

	config.ApiKey = key

	return config, nil
}
