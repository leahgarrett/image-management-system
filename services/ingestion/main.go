package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/api"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/config"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/jobs"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	s3Client, err := storage.NewS3Client(storage.Config{
		Region: cfg.AWSRegion,
		Bucket: cfg.S3Bucket,
	})
	if err != nil {
		log.Fatalf("S3 client: %v", err)
	}

	tmpDir := os.TempDir()
	store := jobs.NewStore()
	pool := processor.NewWorkerPool(cfg.WorkerCount, s3Client)
	handlers := api.NewHandlers(store, pool, cfg.MaxFileSizeBytes, tmpDir)
	router := api.NewRouter(handlers, cfg.JWTSecret)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("ingestion service starting on :%s (workers: %d, max upload: %dMB)",
		cfg.Port, cfg.WorkerCount, cfg.MaxFileSizeBytes/1024/1024)
	log.Fatal(srv.ListenAndServe())
}
