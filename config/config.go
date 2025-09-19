package config

import (
    "os"
    "strconv"
    "time"
    
    "github.com/joho/godotenv"
    "github.com/sirupsen/logrus"
)

type Config struct {
    AdsAPIURL     string
    CRMAPIURL     string
    SinkURL       string
    SinkSecret    string
    Port          string
    LogLevel      string
    HTTPTimeout   time.Duration
    RetryAttempts int
}

func Load() *Config {
    // Load .env file if it exists
    if err := godotenv.Load(); err != nil {
        logrus.Warn("No .env file found, using environment variables")
    }

    timeout, _ := time.ParseDuration(getEnv("HTTP_TIMEOUT", "30s"))
    retryAttempts, _ := strconv.Atoi(getEnv("RETRY_ATTEMPTS", "3"))

    return &Config{
        AdsAPIURL:     getEnv("ADS_API_URL", "https://mocki.io/v1/9dcc2981-2bc8-465a-bce3-47767e1278e6"),
        CRMAPIURL:     getEnv("CRM_API_URL", "https://mocki.io/v1/6a064f10-829d-432c-9f0d-24d5b8cb71c7"),
        SinkURL:       getEnv("SINK_URL", "https://httpbin.org/post"),
        SinkSecret:    getEnv("SINK_SECRET", "admira_secret_example"),
        Port:          getEnv("PORT", "8080"),
        LogLevel:      getEnv("LOG_LEVEL", "info"),
        HTTPTimeout:   timeout,
        RetryAttempts: retryAttempts,
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
