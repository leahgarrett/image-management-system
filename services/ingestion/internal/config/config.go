package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port             string
	WorkerCount      int
	MaxFileSizeBytes int64
	JWTSecret        string
	AWSRegion        string
	S3Bucket         string
}

func Load() (*Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}

	port := getEnvOrDefault("PORT", "8080")

	workerCount := 10
	if v := os.Getenv("WORKER_COUNT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("WORKER_COUNT must be a positive integer")
		}
		workerCount = n
	}

	maxFileSizeMB := int64(15)
	if v := os.Getenv("MAX_FILE_SIZE_MB"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("MAX_FILE_SIZE_MB must be a positive integer")
		}
		maxFileSizeMB = n
	}

	return &Config{
		Port:             port,
		WorkerCount:      workerCount,
		MaxFileSizeBytes: maxFileSizeMB * 1024 * 1024,
		JWTSecret:        jwtSecret,
		AWSRegion:        awsRegion,
		S3Bucket:         s3Bucket,
	}, nil
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
