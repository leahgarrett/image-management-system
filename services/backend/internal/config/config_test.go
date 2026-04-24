package config_test

import (
	"testing"

	"github.com/leahgarrett/image-management-system/services/backend/internal/config"
)

func setBase(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://backend:backend@localhost:5432/imagedb?sslmode=disable")
	t.Setenv("JWT_SECRET", "testsecret")
	t.Setenv("APP_URL", "http://localhost:3000")
	t.Setenv("DEV_MODE", "true")
}

func TestLoad_Defaults(t *testing.T) {
	setBase(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8081" {
		t.Errorf("expected port 8081, got %s", cfg.Port)
	}
	if cfg.SMTPPort != 587 {
		t.Errorf("expected SMTP port 587, got %d", cfg.SMTPPort)
	}
	if cfg.DevMode != true {
		t.Error("expected DevMode true")
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "testsecret")
	t.Setenv("APP_URL", "http://localhost:3000")
	t.Setenv("DEV_MODE", "true")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://backend:backend@localhost:5432/imagedb?sslmode=disable")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("APP_URL", "http://localhost:3000")
	t.Setenv("DEV_MODE", "true")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET")
	}
}

func TestLoad_SMTPRequiredWhenNotDevMode(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://backend:backend@localhost:5432/imagedb?sslmode=disable")
	t.Setenv("JWT_SECRET", "testsecret")
	t.Setenv("APP_URL", "http://localhost:3000")
	t.Setenv("DEV_MODE", "false")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_FROM", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing SMTP config when DEV_MODE=false")
	}
}

func TestLoad_SMTPNotRequiredInDevMode(t *testing.T) {
	setBase(t)
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_FROM", "")
	_, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error in dev mode without SMTP: %v", err)
	}
}
