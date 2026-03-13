package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	CORS     CORSConfig
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type ServerConfig struct {
	Port string
}

type CORSConfig struct {
	AllowOrigins []string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			DSN:             getEnv("POSTGRES_DSN", ""),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: time.Hour,
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		CORS: CORSConfig{
			AllowOrigins: getEnvStringSlice(
				"CORS_ALLOW_ORIGINS",
				[]string{"http://localhost:5173", "http://localhost:4173"},
			),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
