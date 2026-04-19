package processor_test

import (
	"image/jpeg"
	"os"
	"testing"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
)

func longestSide(t *testing.T, path string) int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	cfg, err := jpeg.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decode config %s: %v", path, err)
	}
	if cfg.Width >= cfg.Height {
		return cfg.Width
	}
	return cfg.Height
}

func TestGenerateVariants_LandscapeThumbnail(t *testing.T) {
	src := writeTempJPEG(t, 4032, 3024)
	result, err := processor.GenerateVariants(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := longestSide(t, result.ThumbnailPath); got != 300 {
		t.Errorf("thumbnail longest side = %d, want 300", got)
	}
}

func TestGenerateVariants_LandscapeWeb(t *testing.T) {
	src := writeTempJPEG(t, 4032, 3024)
	result, err := processor.GenerateVariants(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := longestSide(t, result.WebPath); got != 1920 {
		t.Errorf("web longest side = %d, want 1920", got)
	}
}

func TestGenerateVariants_PortraitThumbnail(t *testing.T) {
	src := writeTempJPEG(t, 3024, 4032)
	result, err := processor.GenerateVariants(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := longestSide(t, result.ThumbnailPath); got != 300 {
		t.Errorf("portrait thumbnail longest side = %d, want 300", got)
	}
}

func TestGenerateVariants_NoUpscale(t *testing.T) {
	src := writeTempJPEG(t, 200, 150)
	result, err := processor.GenerateVariants(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := longestSide(t, result.ThumbnailPath); got > 200 {
		t.Errorf("upscaled thumbnail to %d, should not exceed original 200px", got)
	}
}

func TestGenerateVariants_FileSizesPopulated(t *testing.T) {
	src := writeTempJPEG(t, 2000, 1500)
	result, err := processor.GenerateVariants(src, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if result.ThumbnailSize <= 0 {
		t.Error("ThumbnailSize should be > 0")
	}
	if result.WebSize <= 0 {
		t.Error("WebSize should be > 0")
	}
}
