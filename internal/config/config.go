package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	Onboarding OnboardingConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Address string
	Timeout time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// OnboardingConfig holds onboarding-specific configuration
type OnboardingConfig struct {
	MaxRetries      int
	RetryDelay      time.Duration
	SessionTimeout  time.Duration
	ValidationRules string // Path to validation rules file
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Address: getEnv("SERVER_ADDRESS", ":8080"),
			Timeout: getDurationEnv("SERVER_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getIntEnv("DB_PORT", 5432),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getIntEnv("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Onboarding: OnboardingConfig{
			MaxRetries:      getIntEnv("ONBOARDING_MAX_RETRIES", 3),
			RetryDelay:      getDurationEnv("ONBOARDING_RETRY_DELAY", 5*time.Second),
			SessionTimeout:  getDurationEnv("ONBOARDING_SESSION_TIMEOUT", 24*time.Hour),
			ValidationRules: getEnv("VALIDATION_RULES_PATH", "./config/validation_rules.yaml"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
