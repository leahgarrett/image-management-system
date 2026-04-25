package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/jobs"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
)

type Handlers struct {
	store        *jobs.Store
	pool         *processor.WorkerPool
	maxFileBytes int64
	tmpDir       string
}

func NewHandlers(store *jobs.Store, pool *processor.WorkerPool, maxFileBytes int64, tmpDir string) *Handlers {
	return &Handlers{store: store, pool: pool, maxFileBytes: maxFileBytes, tmpDir: tmpDir}
}

func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, errUnauthorized("missing user context"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxFileBytes)
	if err := r.ParseMultipartForm(h.maxFileBytes); err != nil {
		writeError(w, r, errTooLarge("file exceeds maximum allowed size"))
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, r, errValidation("missing 'image' field in multipart form"))
		return
	}
	defer file.Close()

	if header.Size > h.maxFileBytes {
		writeError(w, r, errTooLarge("file exceeds maximum allowed size"))
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".heic": true, ".heif": true, ".tiff": true, ".bmp": true, ".webp": true}
	if !allowed[ext] {
		writeError(w, r, errValidation("unsupported file type: "+ext))
		return
	}

	imageID := uuid.NewString()
	jobOutDir := filepath.Join(h.tmpDir, imageID)
	if err := os.MkdirAll(jobOutDir, 0700); err != nil {
		writeError(w, r, errInternal("could not create working directory"))
		return
	}

	submitted := false
	defer func() {
		if !submitted {
			os.RemoveAll(jobOutDir)
		}
	}()

	tmpPath := filepath.Join(jobOutDir, "upload"+ext)
	dst, err := os.Create(tmpPath)
	if err != nil {
		writeError(w, r, errInternal("could not store upload"))
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		writeError(w, r, errInternal("could not store upload"))
		return
	}
	dst.Close()

	job := h.store.Create(imageID, userID, header.Filename)

	uploadJob := processor.UploadJob{
		ImageID:      imageID,
		UserID:       userID,
		FilePath:     tmpPath,
		OriginalName: header.Filename,
		OutDir:       jobOutDir,
	}

	h.pool.Submit(r.Context(), uploadJob, func(result processor.ProcessResult) {
		defer os.RemoveAll(jobOutDir)
		if result.Error != nil {
			h.store.SetFailed(job.ID, result.Error.Error())
			return
		}
		h.store.SetCompleted(job.ID, jobs.CompletedResult{
			ThumbnailKey: result.ThumbnailKey,
			WebKey:       result.WebKey,
			OriginalKey:  result.OriginalKey,
			Metadata:     exifToMap(result.Metadata),
		})
	})
	submitted = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"jobId":   job.ID,
		"imageId": imageID,
		"status":  "queued",
	})
}

func (h *Handlers) Status(w http.ResponseWriter, r *http.Request, jobID string) {
	job, ok := h.store.Get(jobID)
	if !ok {
		writeError(w, r, errNotFound("job not found"))
		return
	}

	resp := map[string]any{
		"jobId":   job.ID,
		"imageId": job.ImageID,
		"status":  job.Status,
	}
	if job.Stage != "" {
		resp["stage"] = job.Stage
	}
	if job.Status == jobs.StatusCompleted {
		resp["keys"] = map[string]string{
			"thumbnail": job.ThumbnailKey,
			"web":       job.WebKey,
			"original":  job.OriginalKey,
		}
		resp["metadata"] = job.Metadata
	}
	if job.Status == jobs.StatusFailed {
		resp["error"] = job.ErrorMessage
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func exifToMap(data processor.EXIFData) map[string]any {
	m := map[string]any{}
	if data.CaptureDate != nil {
		m["captureDate"] = data.CaptureDate
	}
	if data.CameraMake != "" {
		m["cameraMake"] = data.CameraMake
	}
	if data.CameraModel != "" {
		m["cameraModel"] = data.CameraModel
	}
	if data.Width > 0 {
		m["width"] = data.Width
	}
	if data.Height > 0 {
		m["height"] = data.Height
	}
	if data.Orientation > 0 {
		m["orientation"] = data.Orientation
	}
	return m
}
