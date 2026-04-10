package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port              string
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	KafkaBrokers      []string
	KafkaTopic        string
	DBMaxRetries      int
	DBRetryDelay      time.Duration
	KafkaRetryDelay   time.Duration
	ProjectionEnabled bool
}

func Load() Config {
	return Config{
		Port:              getEnv("API_PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		KafkaBrokers:      splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:        getEnv("KAFKA_TOPIC", "user-events"),
		DBMaxRetries:      getEnvInt("DB_MAX_RETRIES", 30),
		DBRetryDelay:      time.Duration(getEnvInt("DB_RETRY_DELAY", 1)) * time.Second,
		KafkaRetryDelay:   time.Duration(getEnvInt("KAFKA_RETRY_DELAY", 2)) * time.Second,
		ProjectionEnabled: getEnvBool("PROJECTION_ENABLED", true),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes"
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
