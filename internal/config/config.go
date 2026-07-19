package config
import (
	"os"
	"encoding/json"
	"github.com/joho/godotenv"
	"errors"
	"fmt"
)

type Config struct {
	ServerAddress string `json:"serverAddress"`
	RedisAddress string `json:"redisAddress"`
	ApiUrl string `json:"visualCrossingApiUrl"`
	ExpirationTime int `json:"expirationTimeInHours"`
	ApiKey string
}

func LoadConfig() (*Config, error) {
	var config *Config

	err := godotenv.Load()
	if err != nil {
		return nil,
		fmt.Errorf("could not load environment variables: %w. Please make sure that the .env file is located in the project root and mathes the .env.example file",
		err)
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

