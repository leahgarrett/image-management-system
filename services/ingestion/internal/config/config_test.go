package config_test

import (
	"os"
	"testing"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/config"
)

func setEnv(t *testing.T, pairs ...string) {
	t.Helper()
	for i := 0; i < len(pairs); i += 2 {
		k := pairs[i]
		v := pairs[i+1]
		os.Setenv(k, v)
		t.Cleanup(func() { os.Unsetenv(k) })
	}
}

func TestLoad_Defaults(t *testing.T) {
	setEnv(t, "JWT_SECRET", "secret", "AWS_REGION", "ap-southeast-2", "S3_BUCKET", "bucket")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.WorkerCount != 10 {
		t.Errorf("WorkerCount = %d, want 10", cfg.WorkerCount)
	}
	if cfg.MaxFileSizeBytes != 15*1024*1024 {
		t.Errorf("MaxFileSizeBytes = %d, want %d", cfg.MaxFileSizeBytes, 15*1024*1024)
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	setEnv(t, "AWS_REGION", "ap-southeast-2", "S3_BUCKET", "bucket")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET")
	}
}

func TestLoad_MissingAWSRegion(t *testing.T) {
	setEnv(t, "JWT_SECRET", "secret", "S3_BUCKET", "bucket")
	os.Unsetenv("AWS_REGION")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing AWS_REGION")
	}
}

func TestLoad_MissingS3Bucket(t *testing.T) {
	setEnv(t, "JWT_SECRET", "secret", "AWS_REGION", "ap-southeast-2")
	os.Unsetenv("S3_BUCKET")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing S3_BUCKET")
	}
}

func TestLoad_CustomWorkerCount(t *testing.T) {
	setEnv(t, "JWT_SECRET", "s", "AWS_REGION", "r", "S3_BUCKET", "b", "WORKER_COUNT", "5")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WorkerCount != 5 {
		t.Errorf("WorkerCount = %d, want 5", cfg.WorkerCount)
	}
}
