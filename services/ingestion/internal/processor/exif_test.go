package processor_test

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"testing"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
)

// writeTempJPEG is a shared helper used by multiple test files in this package.
func writeTempJPEG(t *testing.T, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 128, G: 64, B: 32, A: 255})
		}
	}
	f, err := os.CreateTemp("", "test-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestExtractEXIF_SyntheticJPEG_NoData(t *testing.T) {
	path := writeTempJPEG(t, 800, 600)

	data, err := processor.ExtractEXIF(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.CaptureDate != nil {
		t.Error("expected nil CaptureDate for synthetic JPEG")
	}
	if data.CameraMake != "" {
		t.Errorf("CameraMake = %q, want empty", data.CameraMake)
	}
	if data.CameraModel != "" {
		t.Errorf("CameraModel = %q, want empty", data.CameraModel)
	}
}

func TestExtractEXIF_GPSAlwaysZero(t *testing.T) {
	path := writeTempJPEG(t, 400, 300)
	data, err := processor.ExtractEXIF(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.GPSLatitude != 0 || data.GPSLongitude != 0 {
		t.Error("GPS coordinates must always be zero (stripped for privacy)")
	}
}

func TestExtractEXIF_MissingFile(t *testing.T) {
	_, err := processor.ExtractEXIF("/tmp/does-not-exist-12345.jpg")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
