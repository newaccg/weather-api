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
	ApiKey string

	ServerAddress string `json:"serverAddress"`
	ApiUrl        string `json:"visualCrossingApiUrl"`

	Timeouts   Timeouts            `json:"timeouts"`
	RateLimits RateLimitsPerMinute `json:"rateLimitsPerMinute"`
	DB         DBConfig            `json:"database"`
}

type DBConfig struct {
	Password string

	KeyExpirationTime Duration `json:"keyExpirationTime"`
	Address           string   `json:"address"`
	DatabaseNumber    int      `json:"databaseNumber"`
}

type Timeouts struct {
	Weather Duration `json:"weather"`
	Health  Duration `json:"health"`
}

type RateLimitsPerMinute struct {
	Weather int `json:"weather"`
	Health  int `json:"health"`
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

func LoadConfig(configPath string) (*Config, error) {
	var config *Config

	err := godotenv.Load()
	if err != nil {
		return nil,
			fmt.Errorf("could not load environment variables: %w. Please make sure that the .env file is located in the project root and mathes the .env.example file",
				err,
			)
	}

	js, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config.json: %w", err)
	}

	err = json.Unmarshal(js, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config.json: %w", err)
	}

	config.ApiKey = os.Getenv("VISUAL_CROSSING_API_KEY")
	if config.ApiKey == "" {
		return nil, errors.New("Visual Crossing API key is not specified")
	}

	config.DB.Password = os.Getenv("DB_PASSWORD")

	return config, nil
}
