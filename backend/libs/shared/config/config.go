package config

import (
	"fmt"
	"os"
	"strconv"
)

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

type JWTConfig struct {
	Secret           string
	AccessTTLMinutes int
}

type BootstrapAdminConfig struct {
	Email    string
	Password string
}

type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	JWT            JWTConfig
	BootstrapAdmin BootstrapAdminConfig
}

func Load() (Config, error) {
	port, err := envInt("PORT", 8080)
	if err != nil {
		return Config{}, fmt.Errorf("invalid PORT: %w", err)
	}

	dbPort, err := envInt("DB_PORT", 5432)
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	jwtTTL, err := envInt("JWT_ACCESS_TTL_MINUTES", 15)
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_ACCESS_TTL_MINUTES: %w", err)
	}

	cfg := Config{
		Server: ServerConfig{Port: port},
		Database: DatabaseConfig{
			Host:     envString("DB_HOST", "localhost"),
			Port:     dbPort,
			Name:     envString("DB_NAME", "vocabulary"),
			User:     envString("DB_USER", "vocabulary"),
			Password: envString("DB_PASSWORD", "vocabulary"),
			SSLMode:  envString("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:           envString("JWT_SECRET", "change-me"),
			AccessTTLMinutes: jwtTTL,
		},
		BootstrapAdmin: BootstrapAdminConfig{
			Email:    envString("BOOTSTRAP_ADMIN_EMAIL", ""),
			Password: envString("BOOTSTRAP_ADMIN_PASSWORD", ""),
		},
		CORSAllowedOrigins: envString("CORS_ALLOWED_ORIGINS", "*"),
	}

	return cfg, nil
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

