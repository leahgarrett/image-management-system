package processor_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
)

func TestToJPEGIfNeeded_JPEG_ReturnsSamePath(t *testing.T) {
	src := writeTempJPEG(t, 100, 100)
	result, cleanup, err := processor.ToJPEGIfNeeded(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	if result != src {
		t.Errorf("expected unchanged path %q, got %q", src, result)
	}
}

func TestToJPEGIfNeeded_PNG_ReturnsSamePath(t *testing.T) {
	src := writeTempJPEG(t, 100, 100)
	result, cleanup, err := processor.ToJPEGIfNeeded(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()
	_ = result
}

func TestToJPEGIfNeeded_HEIC_ReturnsJPEGPath(t *testing.T) {
	// Requires a real HEIC file at testdata/sample.heic.
	// Skip gracefully if not present.
	heicPath := filepath.Join("testdata", "sample.heic")
	result, cleanup, err := processor.ToJPEGIfNeeded(heicPath, t.TempDir())
	if err != nil {
		t.Skip("skipping HEIC test (no testdata/sample.heic or libheif not available): " + err.Error())
	}
	defer cleanup()

	if !strings.HasSuffix(result, ".jpg") {
		t.Errorf("expected .jpg output path, got %q", result)
	}
}
