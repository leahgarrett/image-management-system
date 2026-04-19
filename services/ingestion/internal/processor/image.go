package processor

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

const (
	thumbnailMaxDim  = 300
	webMaxDim        = 1920
	thumbnailQuality = 85
	webQuality       = 90
)

type VariantResult struct {
	ThumbnailPath string
	WebPath       string
	ThumbnailSize int64
	WebSize       int64
}

// GenerateVariants resizes src to thumbnail (300px longest side) and web (1920px longest side)
// JPEG variants, writing output into outDir. Images smaller than the target are not upscaled.
// EXIF orientation is applied automatically before resizing.
func GenerateVariants(src, outDir string) (VariantResult, error) {
	img, err := imaging.Open(src, imaging.AutoOrientation(true))
	if err != nil {
		return VariantResult{}, fmt.Errorf("open %s: %w", src, err)
	}

	thumbPath := filepath.Join(outDir, "thumbnail.jpg")
	if err := saveResized(img, thumbPath, thumbnailMaxDim, thumbnailQuality); err != nil {
		return VariantResult{}, fmt.Errorf("thumbnail: %w", err)
	}

	webPath := filepath.Join(outDir, "web.jpg")
	if err := saveResized(img, webPath, webMaxDim, webQuality); err != nil {
		return VariantResult{}, fmt.Errorf("web: %w", err)
	}

	thumbSize, err := fileSize(thumbPath)
	if err != nil {
		return VariantResult{}, err
	}
	webSize, err := fileSize(webPath)
	if err != nil {
		return VariantResult{}, err
	}

	return VariantResult{
		ThumbnailPath: thumbPath,
		WebPath:       webPath,
		ThumbnailSize: thumbSize,
		WebSize:       webSize,
	}, nil
}

// saveResized resizes img so its longest side is maxDim, then saves as JPEG.
// If both dimensions are already <= maxDim, saves without upscaling.
func saveResized(img image.Image, dst string, maxDim, quality int) error {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	if w <= maxDim && h <= maxDim {
		return imaging.Save(img, dst, imaging.JPEGQuality(quality))
	}

	var resized image.Image
	if w >= h {
		resized = imaging.Resize(img, maxDim, 0, imaging.Lanczos)
	} else {
		resized = imaging.Resize(img, 0, maxDim, imaging.Lanczos)
	}
	return imaging.Save(resized, dst, imaging.JPEGQuality(quality))
}

func fileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
