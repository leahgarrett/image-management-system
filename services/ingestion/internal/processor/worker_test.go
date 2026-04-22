package processor_test

import (
	"context"
	"os"
	"testing"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
)

type stubUploader struct {
	keys []string
}

func (s *stubUploader) Upload(_ context.Context, _, key, _ string) error {
	s.keys = append(s.keys, key)
	return nil
}

func TestWorkerPool_Process_Success(t *testing.T) {
	stub := &stubUploader{}
	pool := processor.NewWorkerPool(2, stub)

	src := writeTempJPEG(t, 1200, 900)
	outDir, err := os.MkdirTemp("", "worker-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outDir)

	job := processor.UploadJob{
		ImageID:      "img-test-001",
		UserID:       "user-001",
		FilePath:     src,
		OriginalName: "photo.jpg",
		OutDir:       outDir,
	}

	result := pool.Process(context.Background(), job)
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.ThumbnailKey == "" {
		t.Error("ThumbnailKey must not be empty")
	}
	if result.WebKey == "" {
		t.Error("WebKey must not be empty")
	}
	if result.OriginalKey == "" {
		t.Error("OriginalKey must not be empty")
	}
	if len(stub.keys) != 3 {
		t.Errorf("expected 3 S3 uploads, got %d: %v", len(stub.keys), stub.keys)
	}
}

func TestWorkerPool_Process_KeyFormat(t *testing.T) {
	stub := &stubUploader{}
	pool := processor.NewWorkerPool(1, stub)

	src := writeTempJPEG(t, 500, 400)
	outDir, _ := os.MkdirTemp("", "worker-key-test-*")
	defer os.RemoveAll(outDir)

	result := pool.Process(context.Background(), processor.UploadJob{
		ImageID:      "img-abc",
		UserID:       "user-xyz",
		FilePath:     src,
		OriginalName: "shot.jpg",
		OutDir:       outDir,
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.ThumbnailKey != "user-xyz/img-abc/thumbnail.jpg" {
		t.Errorf("ThumbnailKey = %q, want %q", result.ThumbnailKey, "user-xyz/img-abc/thumbnail.jpg")
	}
	if result.WebKey != "user-xyz/img-abc/web.jpg" {
		t.Errorf("WebKey = %q, want %q", result.WebKey, "user-xyz/img-abc/web.jpg")
	}
	if result.OriginalKey != "user-xyz/img-abc/original.jpg" {
		t.Errorf("OriginalKey = %q, want %q", result.OriginalKey, "user-xyz/img-abc/original.jpg")
	}
}
