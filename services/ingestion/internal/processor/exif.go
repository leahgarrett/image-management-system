package processor

import (
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type EXIFData struct {
	Width        int
	Height       int
	Orientation  int
	CaptureDate  *time.Time
	CameraMake   string
	CameraModel  string
	// GPSLatitude and GPSLongitude are intentionally never extracted.
	// Fields exist only so callers can check the zero value.
	GPSLatitude  float64
	GPSLongitude float64
}

// ExtractEXIF reads safe EXIF fields from a JPEG file. GPS coordinates and
// device serial numbers are never extracted. Returns empty EXIFData (not an
// error) for files with no EXIF block.
func ExtractEXIF(path string) (EXIFData, error) {
	f, err := os.Open(path)
	if err != nil {
		return EXIFData{}, err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		// No EXIF present — common for synthetic images. Not an error.
		return EXIFData{}, nil
	}

	var data EXIFData

	if dt, err := x.DateTime(); err == nil {
		data.CaptureDate = &dt
	}
	if make_, err := x.Get(exif.Make); err == nil {
		data.CameraMake, _ = make_.StringVal()
	}
	if model, err := x.Get(exif.Model); err == nil {
		data.CameraModel, _ = model.StringVal()
	}
	if w, err := x.Get(exif.PixelXDimension); err == nil {
		data.Width, _ = w.Int(0)
	}
	if h, err := x.Get(exif.PixelYDimension); err == nil {
		data.Height, _ = h.Int(0)
	}
	if o, err := x.Get(exif.Orientation); err == nil {
		data.Orientation, _ = o.Int(0)
	}

	return data, nil
}
