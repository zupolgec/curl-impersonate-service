package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const Version = "1.2.0"

type Config struct {
	Token               string
	Port                string
	LogLevel            string
	MaxRequestBodySize  int64
	MaxResponseBodySize int64
	MaxTimeout          int
	DefaultTimeout      int
	BrowsersJSONPath    string

	// SSRF protection
	SSRFAllowPrivate bool
	SSRFDenyHosts    []string
	SSRFAllowHosts   []string

	// CORS: initial allowed origins, "*" for any. May be overridden at runtime
	// via the admin UI (persisted in the datastore).
	CORSAllowedOrigins []string

	// Persistence and admin UI
	AdminToken        string
	DataDir           string
	LogRetentionHours int
}

func Load() (*Config, error) {
	token := os.Getenv("TOKEN")
	adminToken := os.Getenv("ADMIN_TOKEN")
	if token == "" && adminToken == "" {
		return nil, fmt.Errorf("set TOKEN and/or ADMIN_TOKEN: at least one is required")
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

		SSRFAllowPrivate: getEnvBool("SSRF_ALLOW_PRIVATE", false),
		SSRFDenyHosts:    getEnvList("SSRF_DENY_HOSTS"),
		SSRFAllowHosts:   getEnvList("SSRF_ALLOW_HOSTS"),

		CORSAllowedOrigins: corsOrigins(),

		AdminToken:        adminToken,
		DataDir:           getEnvOrDefault("DATA_DIR", "/data"),
		LogRetentionHours: getEnvIntOrDefault("LOG_RETENTION_HOURS", 72),
	}

	return cfg, nil
}

// corsOrigins reads CORS_ALLOWED_ORIGINS, defaulting to "*" (any origin).
func corsOrigins() []string {
	if origins := getEnvList("CORS_ALLOWED_ORIGINS"); origins != nil {
		return origins
	}
	return []string{"*"}
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// getEnvList parses a comma-separated env var into a trimmed, non-empty slice.
func getEnvList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	var out []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
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
