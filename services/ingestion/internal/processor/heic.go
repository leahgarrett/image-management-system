package processor

import (
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/strukturag/libheif/go/heif"
)

// ToJPEGIfNeeded converts a HEIC/HEIF file to JPEG and returns the output path
// along with a cleanup function. For non-HEIC files it returns the original path
// unchanged and a no-op cleanup. The caller must always call cleanup().
func ToJPEGIfNeeded(src, outDir string) (path string, cleanup func(), err error) {
	noop := func() {}
	ext := strings.ToLower(filepath.Ext(src))
	if ext != ".heic" && ext != ".heif" {
		return src, noop, nil
	}

	ctx, err := heif.NewContext()
	if err != nil {
		return "", noop, fmt.Errorf("heif context: %w", err)
	}

	if err := ctx.ReadFromFile(src); err != nil {
		return "", noop, fmt.Errorf("heif read: %w", err)
	}

	handle, err := ctx.GetPrimaryImageHandle()
	if err != nil {
		return "", noop, fmt.Errorf("heif primary handle: %w", err)
	}

	img, err := handle.DecodeImage(heif.ColorspaceRGB, heif.ChromaInterleavedRGB, nil)
	if err != nil {
		return "", noop, fmt.Errorf("heif decode: %w", err)
	}

	goImg, err := img.GetImage()
	if err != nil {
		return "", noop, fmt.Errorf("heif to Go image: %w", err)
	}

	outPath := filepath.Join(outDir, "converted.jpg")
	f, err := os.Create(outPath)
	if err != nil {
		return "", noop, fmt.Errorf("create output JPEG: %w", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, goImg, &jpeg.Options{Quality: 95}); err != nil {
		return "", noop, fmt.Errorf("jpeg encode: %w", err)
	}

	// outDir is unique per job (UUID-based), so filename collision is not possible.
	return outPath, func() { _ = os.Remove(outPath) }, nil
}
