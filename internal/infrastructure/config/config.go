package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port           string
	LogLevel       string // info, debug, ...
	DatabaseString string
	// Environment information
	Environment string
	Version     string
	Debug       bool

	// Security
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		DatabaseString: getEnv("DATABASE_STRING", "host=localhost port=5432 user=postgres dbname=hackspark password=postgres sslmode=disable"),
		Environment:    getEnv("ENVIRONMENT", "development"),

		// Security defaults
		AllowedOrigins: getSliceEnv("CORS_ALLOWED_ORIGINS", []string{"*"}),
		AllowedMethods: getSliceEnv("CORS_ALLOWED_METHODS", []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		}),
		AllowedHeaders: getSliceEnv("CORS_ALLOWED_HEADERS", []string{
			"Accept", "Authorization", "Content-Type", "X-CSRF-Token",
		}),
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	// Validate environment
	validEnvironments := map[string]bool{
		"development": true,
		"test":        true,
		"staging":     true,
		"production":  true,
	}

	if !validEnvironments[c.Environment] {
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
		"trace": true,
	}

	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	return nil
}

// Helper functions to read environment variables
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}
	return defaultValue
}
