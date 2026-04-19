package main

import (
	"log"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("ingestion service will start on :%s (workers: %d)", cfg.Port, cfg.WorkerCount)
}
