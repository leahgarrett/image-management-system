package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	AppURL      string
	DevMode     bool
	SMTPHost    string
	SMTPPort    int
	SMTPFrom    string
}

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		return nil, fmt.Errorf("APP_URL is required")
	}

	port := getEnvOrDefault("PORT", "8081")

	devMode := false
	if v := os.Getenv("DEV_MODE"); v == "true" {
		devMode = true
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpFrom := os.Getenv("SMTP_FROM")
	smtpPort := 587

	if !devMode {
		if smtpHost == "" {
			return nil, fmt.Errorf("SMTP_HOST is required when DEV_MODE is not true")
		}
		if smtpFrom == "" {
			return nil, fmt.Errorf("SMTP_FROM is required when DEV_MODE is not true")
		}
		if v := os.Getenv("SMTP_PORT"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				return nil, fmt.Errorf("SMTP_PORT must be a positive integer")
			}
			smtpPort = n
		}
	}

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		AppURL:      appURL,
		DevMode:     devMode,
		SMTPHost:    smtpHost,
		SMTPPort:    smtpPort,
		SMTPFrom:    smtpFrom,
	}, nil
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
