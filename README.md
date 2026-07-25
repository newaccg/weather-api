# Golang Weather API

Project idea is taken from [Roadmap.sh](https://roadmap.sh/projects/weather-api-wrapper-service).

This API uses third-party API (Visual Crossing) for weather data

## Features

* Redis **caching**

* **Clean** Go **architecture**

* **Configuration** via .env (using [godotenv](https://github.com/joho/godotenv)) file and internal/config/config.json

* **Handling** errors, such as external API errors, network errors, etc

* Structured **logging**

## Requirements

* Working Redis server

* Go 1.24+

## Endpoints

### Health

```textile
GET /health
```

Returns 200 if healthy, otherwise returns 503 and a list of non-working components via JSON



```textile
GET /weather/{city}
```

Returns weather by city via JSON

## Getting started

1. **Clone this repository**

```bash
git clone https://github.com/newaccg/weather-api.git
```

2. **Go to the cloned repository**

```bash
cd weather-api
```

3. **Install dependencies**

```bash
go mod download
```

4. **Create .env file and fill it according to .env.example file**

```textile
VISUAL_CROSSING_API_KEY="YOUR_API_KEY"
```

5.  (optional) **edit internal/config/config.json**

```bash
nano internal/config/config.json
```

6. **Run the application**

```bash
go run cmd/main.go
```

7. Or, if you prefer, **build and run this project for maximum prefomance**

```bash
go build -ldflags="-s -w" -o weather-api cmd/main.go
./weather-api
```

## Structure

```textile
.
├── .env # secrets
├── .env.example # example of .env
├── README.md
├── cmd
│   └── main.go # entry point
├── go.mod
├── go.sum
└── internal
    ├── client
    │   └── client.go # redefining HTTP client
    ├── config
    │   ├── config.go # configuration loading
    │   └── config.json # general configuration
    ├── errors
    │   └── errors.go # custom error structures and methods
    ├── handler
    │   └── handler.go
    ├── middleware
    │   └── middleware.go # limiting
    ├── repository
    │   └── repository.go # redis caching
    └── service
        └── service.go # logic
```

