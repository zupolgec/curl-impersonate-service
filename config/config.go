package config

import (
	"fmt"
	"os"
	"strconv"
)

const Version = "1.0.0"

type Config struct {
	Token               string
	Port                string
	LogLevel            string
	MaxRequestBodySize  int64
	MaxResponseBodySize int64
	MaxTimeout          int
	DefaultTimeout      int
	BrowsersJSONPath    string
}

func Load() (*Config, error) {
	token := os.Getenv("TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TOKEN environment variable is required")
	}

	cfg := &Config{
		Token:               token,
		Port:                getEnvOrDefault("PORT", "8080"),
		LogLevel:            getEnvOrDefault("LOG_LEVEL", "info"),
		MaxRequestBodySize:  getEnvInt64OrDefault("MAX_REQUEST_BODY_SIZE", 10485760),  // 10MB
		MaxResponseBodySize: getEnvInt64OrDefault("MAX_RESPONSE_BODY_SIZE", 52428800), // 50MB
		MaxTimeout:          getEnvIntOrDefault("MAX_TIMEOUT", 120),
		DefaultTimeout:      getEnvIntOrDefault("DEFAULT_TIMEOUT", 30),
		BrowsersJSONPath:    getEnvOrDefault("BROWSERS_JSON_PATH", "/etc/impersonate/browsers.json"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64OrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
