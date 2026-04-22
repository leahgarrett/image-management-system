package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	apiinternal "github.com/leahgarrett/image-management-system/services/ingestion/internal/api"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/jobs"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
)

// minimalJPEG is a valid 1x1 JPEG (smallest valid JPEG).
var minimalJPEG = []byte{
	0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01,
	0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43,
	0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
	0x09, 0x08, 0x0a, 0x0c, 0x14, 0x0d, 0x0c, 0x0b, 0x0b, 0x0c, 0x19, 0x12,
	0x13, 0x0f, 0x14, 0x1d, 0x1a, 0x1f, 0x1e, 0x1d, 0x1a, 0x1c, 0x1c, 0x20,
	0x24, 0x2e, 0x27, 0x20, 0x22, 0x2c, 0x23, 0x1c, 0x1c, 0x28, 0x37, 0x29,
	0x2c, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1f, 0x27, 0x39, 0x3d, 0x38, 0x32,
	0x3c, 0x2e, 0x33, 0x34, 0x32, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
	0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4, 0x00, 0x1f, 0x00, 0x00,
	0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	0x09, 0x0a, 0x0b, 0xff, 0xc4, 0x00, 0xb5, 0x10, 0x00, 0x02, 0x01, 0x03,
	0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7d,
	0xff, 0xda, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3f, 0x00, 0xfb, 0xd2,
	0x8a, 0x28, 0x03, 0xff, 0xd9,
}

func buildUploadRequest(t *testing.T, filename string, body []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("image", filename)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(body)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	// Inject userId as if JWT middleware ran
	ctx := apiinternal.ContextWithUserID(req.Context(), "user-test-001")
	return req.WithContext(ctx)
}

type noopUploader struct{}

func (n *noopUploader) Upload(_ context.Context, _, _, _ string) error { return nil }

func TestHandleUpload_AcceptsValidJPEG(t *testing.T) {
	store := jobs.NewStore()
	pool := processor.NewWorkerPool(1, &noopUploader{})
	h := apiinternal.NewHandlers(store, pool, 15*1024*1024, t.TempDir())

	req := buildUploadRequest(t, "photo.jpg", minimalJPEG)
	rec := httptest.NewRecorder()
	h.Upload(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202", rec.Code)
	}

	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body)
	if body["jobId"] == "" {
		t.Error("expected non-empty jobId in response")
	}
	if body["status"] != "queued" {
		t.Errorf("status = %v, want queued", body["status"])
	}
}

func TestHandleUpload_RejectsOversizedFile(t *testing.T) {
	store := jobs.NewStore()
	pool := processor.NewWorkerPool(1, &noopUploader{})
	h := apiinternal.NewHandlers(store, pool, 10, t.TempDir()) // max 10 bytes

	req := buildUploadRequest(t, "photo.jpg", minimalJPEG) // minimalJPEG > 10 bytes
	rec := httptest.NewRecorder()
	h.Upload(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", rec.Code)
	}
}

func TestHandleStatus_ReturnsJobStatus(t *testing.T) {
	store := jobs.NewStore()
	pool := processor.NewWorkerPool(1, &noopUploader{})
	h := apiinternal.NewHandlers(store, pool, 15*1024*1024, t.TempDir())

	job := store.Create("img-001", "user-001", "photo.jpg")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ingest/status/"+job.ID, nil)
	rec := httptest.NewRecorder()
	h.Status(rec, req, job.ID)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body)
	if body["jobId"] != job.ID {
		t.Errorf("jobId = %v, want %s", body["jobId"], job.ID)
	}
}

func TestHandleStatus_UnknownJob_Returns404(t *testing.T) {
	store := jobs.NewStore()
	pool := processor.NewWorkerPool(1, &noopUploader{})
	h := apiinternal.NewHandlers(store, pool, 15*1024*1024, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ingest/status/unknown", nil)
	rec := httptest.NewRecorder()
	h.Status(rec, req, "unknown")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHandleHealth_Returns200(t *testing.T) {
	store := jobs.NewStore()
	pool := processor.NewWorkerPool(1, &noopUploader{})
	h := apiinternal.NewHandlers(store, pool, 15*1024*1024, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
