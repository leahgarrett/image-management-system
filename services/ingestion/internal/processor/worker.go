package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/storage"
)

// Uploader is satisfied by *storage.S3Client and by test stubs.
type Uploader interface {
	Upload(ctx context.Context, localPath, key, storageClass string) error
}

type UploadJob struct {
	ImageID      string
	UserID       string
	FilePath     string
	OriginalName string
	OutDir       string
}

type ProcessResult struct {
	ImageID      string
	ThumbnailKey string
	WebKey       string
	OriginalKey  string
	Metadata     EXIFData
	Error        error
}

// WorkerPool limits concurrency via a semaphore (buffered channel).
type WorkerPool struct {
	sem      chan struct{}
	uploader Uploader
}

// NewWorkerPool creates a pool that processes at most workers jobs concurrently.
func NewWorkerPool(workers int, uploader Uploader) *WorkerPool {
	return &WorkerPool{
		sem:      make(chan struct{}, workers),
		uploader: uploader,
	}
}

// Process runs a job synchronously, blocking until a worker slot is free.
// Safe to call from multiple goroutines.
func (p *WorkerPool) Process(ctx context.Context, job UploadJob) ProcessResult {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()
	return p.process(ctx, job)
}

// Submit dispatches a job to a goroutine and calls onDone with the result.
func (p *WorkerPool) Submit(ctx context.Context, job UploadJob, onDone func(ProcessResult)) {
	go func() {
		onDone(p.Process(ctx, job))
	}()
}

func (p *WorkerPool) process(ctx context.Context, job UploadJob) ProcessResult {
	result := ProcessResult{ImageID: job.ImageID}

	// Convert HEIC to JPEG if needed (no-op for JPEG/PNG)
	workPath, cleanHeic, err := ToJPEGIfNeeded(job.FilePath, job.OutDir)
	if err != nil {
		result.Error = fmt.Errorf("heic conversion: %w", err)
		return result
	}
	defer cleanHeic()

	// Extract EXIF before resizing (imaging strips EXIF during resize)
	result.Metadata, err = ExtractEXIF(workPath)
	if err != nil {
		result.Error = fmt.Errorf("exif: %w", err)
		return result
	}

	// Resize to thumbnail and web variants
	variants, err := GenerateVariants(workPath, job.OutDir)
	if err != nil {
		result.Error = fmt.Errorf("variants: %w", err)
		return result
	}
	defer os.Remove(variants.ThumbnailPath)
	defer os.Remove(variants.WebPath)

	// Build S3 key prefix: {userId}/{imageId}/
	prefix := job.UserID + "/" + job.ImageID

	ext := filepath.Ext(job.OriginalName)
	if ext == "" {
		ext = ".jpg"
	}

	thumbKey := prefix + "/thumbnail.jpg"
	webKey := prefix + "/web.jpg"
	origKey := prefix + "/original" + ext

	if err := p.uploader.Upload(ctx, variants.ThumbnailPath, thumbKey, storage.StorageClassFor("thumbnail")); err != nil {
		result.Error = fmt.Errorf("upload thumbnail: %w", err)
		return result
	}
	if err := p.uploader.Upload(ctx, variants.WebPath, webKey, storage.StorageClassFor("web")); err != nil {
		result.Error = fmt.Errorf("upload web: %w", err)
		return result
	}
	if err := p.uploader.Upload(ctx, job.FilePath, origKey, storage.StorageClassFor("original")); err != nil {
		result.Error = fmt.Errorf("upload original: %w", err)
		return result
	}

	result.ThumbnailKey = thumbKey
	result.WebKey = webKey
	result.OriginalKey = origKey
	return result
}
