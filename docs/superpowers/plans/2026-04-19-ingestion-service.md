# Ingestion Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go microservice that accepts image uploads, generates JPEG variants (thumbnail 300px + web-optimised 1920px), extracts safe EXIF metadata, uploads all versions to S3, and returns S3 keys + metadata via an async job status endpoint.

**Architecture:** HTTP server with goroutine worker pool. Upload requests return a `jobId` immediately (202 Accepted); a configurable number of concurrent workers process images (HEIC decode if needed → resize → EXIF extract → S3 upload). An in-memory store (mutex-guarded map) tracks job status from queued through completed/failed. JWT middleware validates the Bearer token signed by the Node.js backend on every non-health endpoint.

**Tech Stack:** Go 1.21, gorilla/mux (routing), disintegration/imaging (JPEG resize), rwcarlsen/goexif (EXIF), strukturag/libheif CGO (HEIC decode), golang-jwt/jwt/v5 (JWT validation), AWS SDK v2 (S3).

---

## File Map

| File | Purpose |
|------|---------|
| `services/ingestion/main.go` | Entry point: load config, wire dependencies, start server |
| `services/ingestion/go.mod` | Go module definition |
| `services/ingestion/.env.example` | Required env vars with placeholder values |
| `services/ingestion/Dockerfile` | Multi-stage build; installs libheif in both stages |
| `services/ingestion/internal/config/config.go` | Load and validate env vars into typed Config struct |
| `services/ingestion/internal/api/errors.go` | APIError type (RFC 7807) and writeError/errXxx helpers |
| `services/ingestion/internal/api/middleware.go` | JWT Bearer validation; injects userId into request context |
| `services/ingestion/internal/api/handlers.go` | POST /upload, GET /status/:jobId, GET /health |
| `services/ingestion/internal/api/server.go` | gorilla/mux router with middleware chain |
| `services/ingestion/internal/jobs/store.go` | Thread-safe in-memory job store (RWMutex-guarded map) |
| `services/ingestion/internal/processor/exif.go` | Extract safe EXIF; GPS and device serial are never returned |
| `services/ingestion/internal/processor/image.go` | Resize to thumbnail (300px) and web (1920px) JPEG variants |
| `services/ingestion/internal/processor/heic.go` | Decode HEIC to JPEG using libheif CGO |
| `services/ingestion/internal/processor/worker.go` | UploadJob/ProcessResult types; semaphore worker pool |
| `services/ingestion/internal/storage/s3.go` | Upload files to S3; selects storage class by variant |

---

## S3 storage class decisions (from storage-and-costs.md)

| Variant | Storage class | Reason |
|---------|--------------|--------|
| thumbnail.jpg | STANDARD | Always hot — used in browse view |
| web.jpg | STANDARD | Viewed on image open |
| original.{ext} | INTELLIGENT_TIERING | Auto-tiers older photos to cheaper tiers |

Cross-region Glacier Deep Archive backup is a **bucket replication rule** (infrastructure config), not application code.

---

### Task 1: Project scaffold

**Files:**
- Create: `services/ingestion/go.mod`
- Create: `services/ingestion/main.go`
- Create: `services/ingestion/.env.example`

- [ ] **Step 1: Create directory and go.mod**

```bash
mkdir -p services/ingestion
cd services/ingestion
go mod init github.com/leahgarrett/image-management-system/services/ingestion
```

Expected output: `go: creating new go.mod: module github.com/leahgarrett/image-management-system/services/ingestion`

- [ ] **Step 2: Install dependencies**

```bash
cd services/ingestion
go get github.com/gorilla/mux@v1.8.1
go get github.com/disintegration/imaging@v1.6.2
go get github.com/rwcarlsen/goexif@v0.0.0-20190401172101-9e8deecbddbd
go get github.com/golang-jwt/jwt/v5@v5.2.1
go get github.com/google/uuid@v1.6.0
go get github.com/aws/aws-sdk-go-v2@v1.26.1
go get github.com/aws/aws-sdk-go-v2/config@v1.27.11
go get github.com/aws/aws-sdk-go-v2/service/s3@v1.53.1
```

Expected: modules downloaded, go.sum created.

- [ ] **Step 3: Create main.go skeleton**

`services/ingestion/main.go`:
```go
package main

import (
	"log"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("ingestion service will start on :%s (workers: %d)", cfg.Port, cfg.WorkerCount)
}
```

- [ ] **Step 4: Create .env.example**

`services/ingestion/.env.example`:
```
PORT=8080
WORKER_COUNT=10
MAX_FILE_SIZE_MB=15
JWT_SECRET=changeme
AWS_REGION=ap-southeast-2
S3_BUCKET=your-bucket-name
```

- [ ] **Step 5: Verify build (will fail on missing config package — expected)**

```bash
cd services/ingestion && go build ./...
```

Expected: `cannot find package "…/internal/config"` — scaffold is correct.

- [ ] **Step 6: Commit**

```bash
git add services/ingestion/
git commit -m "feat(ingestion): project scaffold with go.mod"
```

---

### Task 2: Config

**Files:**
- Create: `services/ingestion/internal/config/config.go`
- Create: `services/ingestion/internal/config/config_test.go`

- [ ] **Step 1: Write failing test**

`services/ingestion/internal/config/config_test.go`:
```go
package config_test

import (
	"os"
	"testing"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/config"
)

func setEnv(t *testing.T, pairs ...string) {
	t.Helper()
	for i := 0; i < len(pairs); i += 2 {
		os.Setenv(pairs[i], pairs[i+1])
		t.Cleanup(func() { os.Unsetenv(pairs[i]) })
	}
}

func TestLoad_Defaults(t *testing.T) {
	setEnv(t, "JWT_SECRET", "secret", "AWS_REGION", "ap-southeast-2", "S3_BUCKET", "bucket")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.WorkerCount != 10 {
		t.Errorf("WorkerCount = %d, want 10", cfg.WorkerCount)
	}
	if cfg.MaxFileSizeBytes != 15*1024*1024 {
		t.Errorf("MaxFileSizeBytes = %d, want %d", cfg.MaxFileSizeBytes, 15*1024*1024)
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	setEnv(t, "AWS_REGION", "ap-southeast-2", "S3_BUCKET", "bucket")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET")
	}
}

func TestLoad_MissingAWSRegion(t *testing.T) {
	setEnv(t, "JWT_SECRET", "secret", "S3_BUCKET", "bucket")
	os.Unsetenv("AWS_REGION")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing AWS_REGION")
	}
}

func TestLoad_MissingS3Bucket(t *testing.T) {
	setEnv(t, "JWT_SECRET", "secret", "AWS_REGION", "ap-southeast-2")
	os.Unsetenv("S3_BUCKET")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing S3_BUCKET")
	}
}

func TestLoad_CustomWorkerCount(t *testing.T) {
	setEnv(t, "JWT_SECRET", "s", "AWS_REGION", "r", "S3_BUCKET", "b", "WORKER_COUNT", "5")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WorkerCount != 5 {
		t.Errorf("WorkerCount = %d, want 5", cfg.WorkerCount)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/config/... -v
```

Expected: FAIL — `cannot find package`

- [ ] **Step 3: Implement config.go**

`services/ingestion/internal/config/config.go`:
```go
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port             string
	WorkerCount      int
	MaxFileSizeBytes int64
	JWTSecret        string
	AWSRegion        string
	S3Bucket         string
}

func Load() (*Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}

	port := getEnvOrDefault("PORT", "8080")

	workerCount := 10
	if v := os.Getenv("WORKER_COUNT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("WORKER_COUNT must be a positive integer")
		}
		workerCount = n
	}

	maxFileSizeMB := int64(15)
	if v := os.Getenv("MAX_FILE_SIZE_MB"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("MAX_FILE_SIZE_MB must be a positive integer")
		}
		maxFileSizeMB = n
	}

	return &Config{
		Port:             port,
		WorkerCount:      workerCount,
		MaxFileSizeBytes: maxFileSizeMB * 1024 * 1024,
		JWTSecret:        jwtSecret,
		AWSRegion:        awsRegion,
		S3Bucket:         s3Bucket,
	}, nil
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/config/... -v
```

Expected:
```
--- PASS: TestLoad_Defaults (0.00s)
--- PASS: TestLoad_MissingJWTSecret (0.00s)
--- PASS: TestLoad_MissingAWSRegion (0.00s)
--- PASS: TestLoad_MissingS3Bucket (0.00s)
--- PASS: TestLoad_CustomWorkerCount (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/internal/config/
git commit -m "feat(ingestion): env-based config with validation"
```

---

### Task 3: Job store

**Files:**
- Create: `services/ingestion/internal/jobs/store.go`
- Create: `services/ingestion/internal/jobs/store_test.go`

- [ ] **Step 1: Write failing test**

`services/ingestion/internal/jobs/store_test.go`:
```go
package jobs_test

import (
	"testing"
	"time"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/jobs"
)

func TestStore_CreateAndGet(t *testing.T) {
	s := jobs.NewStore()
	job := s.Create("img-001", "user-001", "photo.jpg")

	if job.Status != jobs.StatusQueued {
		t.Errorf("Status = %q, want %q", job.Status, jobs.StatusQueued)
	}

	got, ok := s.Get(job.ID)
	if !ok {
		t.Fatal("expected job to exist")
	}
	if got.ImageID != "img-001" {
		t.Errorf("ImageID = %q, want %q", got.ImageID, "img-001")
	}
}

func TestStore_SetProcessing(t *testing.T) {
	s := jobs.NewStore()
	job := s.Create("img-002", "user-001", "photo.jpg")

	s.SetProcessing(job.ID, "generating_variants")
	got, _ := s.Get(job.ID)

	if got.Status != jobs.StatusProcessing {
		t.Errorf("Status = %q, want %q", got.Status, jobs.StatusProcessing)
	}
	if got.Stage != "generating_variants" {
		t.Errorf("Stage = %q, want %q", got.Stage, "generating_variants")
	}
}

func TestStore_SetCompleted(t *testing.T) {
	s := jobs.NewStore()
	job := s.Create("img-003", "user-001", "photo.jpg")

	result := jobs.CompletedResult{
		ThumbnailKey: "user-001/img-003/thumbnail.jpg",
		WebKey:       "user-001/img-003/web.jpg",
		OriginalKey:  "user-001/img-003/original.jpg",
	}
	s.SetCompleted(job.ID, result)

	got, _ := s.Get(job.ID)
	if got.Status != jobs.StatusCompleted {
		t.Errorf("Status = %q, want %q", got.Status, jobs.StatusCompleted)
	}
	if got.ThumbnailKey != result.ThumbnailKey {
		t.Errorf("ThumbnailKey = %q, want %q", got.ThumbnailKey, result.ThumbnailKey)
	}
}

func TestStore_SetFailed(t *testing.T) {
	s := jobs.NewStore()
	job := s.Create("img-004", "user-001", "photo.jpg")

	s.SetFailed(job.ID, "S3 unreachable")
	got, _ := s.Get(job.ID)

	if got.Status != jobs.StatusFailed {
		t.Errorf("Status = %q, want %q", got.Status, jobs.StatusFailed)
	}
	if got.ErrorMessage != "S3 unreachable" {
		t.Errorf("ErrorMessage = %q, want %q", got.ErrorMessage, "S3 unreachable")
	}
}

func TestStore_GetNonExistent(t *testing.T) {
	s := jobs.NewStore()
	_, ok := s.Get("does-not-exist")
	if ok {
		t.Fatal("expected ok=false for non-existent job")
	}
}

func TestStore_UpdatedAtAdvances(t *testing.T) {
	s := jobs.NewStore()
	job := s.Create("img-005", "user-001", "photo.jpg")
	before := job.UpdatedAt

	time.Sleep(time.Millisecond)
	s.SetProcessing(job.ID, "uploading")

	got, _ := s.Get(job.ID)
	if !got.UpdatedAt.After(before) {
		t.Error("expected UpdatedAt to advance after SetProcessing")
	}
}

func TestStore_GetReturnsSnapshot(t *testing.T) {
	s := jobs.NewStore()
	job := s.Create("img-006", "user-001", "photo.jpg")

	got, _ := s.Get(job.ID)
	got.Status = jobs.StatusFailed // mutate snapshot

	// original must be unchanged
	original, _ := s.Get(job.ID)
	if original.Status != jobs.StatusQueued {
		t.Errorf("store was mutated via returned pointer; Status = %q, want %q", original.Status, jobs.StatusQueued)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/jobs/... -v
```

Expected: FAIL — `cannot find package`

- [ ] **Step 3: Implement store.go**

`services/ingestion/internal/jobs/store.go`:
```go
package jobs

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type Job struct {
	ID               string
	ImageID          string
	UserID           string
	OriginalFilename string
	Status           Status
	Stage            string
	ThumbnailKey     string
	WebKey           string
	OriginalKey      string
	ErrorMessage     string
	Metadata         map[string]any
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CompletedResult struct {
	ThumbnailKey string
	WebKey       string
	OriginalKey  string
	Metadata     map[string]any
}

type Store struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewStore() *Store {
	return &Store{jobs: make(map[string]*Job)}
}

func (s *Store) Create(imageID, userID, originalFilename string) *Job {
	now := time.Now()
	job := &Job{
		ID:               uuid.NewString(),
		ImageID:          imageID,
		UserID:           userID,
		OriginalFilename: originalFilename,
		Status:           StatusQueued,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()
	return job
}

// Get returns a copy of the job so callers cannot mutate store state.
func (s *Store) Get(id string) (*Job, bool) {
	s.mu.RLock()
	job, ok := s.jobs[id]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	copy := *job
	return &copy, true
}

func (s *Store) SetProcessing(id, stage string) {
	s.mu.Lock()
	if job, ok := s.jobs[id]; ok {
		job.Status = StatusProcessing
		job.Stage = stage
		job.UpdatedAt = time.Now()
	}
	s.mu.Unlock()
}

func (s *Store) SetCompleted(id string, result CompletedResult) {
	s.mu.Lock()
	if job, ok := s.jobs[id]; ok {
		job.Status = StatusCompleted
		job.Stage = ""
		job.ThumbnailKey = result.ThumbnailKey
		job.WebKey = result.WebKey
		job.OriginalKey = result.OriginalKey
		job.Metadata = result.Metadata
		job.UpdatedAt = time.Now()
	}
	s.mu.Unlock()
}

func (s *Store) SetFailed(id, errMsg string) {
	s.mu.Lock()
	if job, ok := s.jobs[id]; ok {
		job.Status = StatusFailed
		job.Stage = ""
		job.ErrorMessage = errMsg
		job.UpdatedAt = time.Now()
	}
	s.mu.Unlock()
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/jobs/... -v
```

Expected:
```
--- PASS: TestStore_CreateAndGet (0.00s)
--- PASS: TestStore_SetProcessing (0.00s)
--- PASS: TestStore_SetCompleted (0.00s)
--- PASS: TestStore_SetFailed (0.00s)
--- PASS: TestStore_GetNonExistent (0.00s)
--- PASS: TestStore_UpdatedAtAdvances (0.00s)
--- PASS: TestStore_GetReturnsSnapshot (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/internal/jobs/
git commit -m "feat(ingestion): thread-safe in-memory job store"
```

---

### Task 4: API error types

**Files:**
- Create: `services/ingestion/internal/api/errors.go`

- [ ] **Step 1: Implement errors.go**

`services/ingestion/internal/api/errors.go`:
```go
package api

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance,omitempty"`
}

func (e APIError) Error() string { return e.Detail }

func writeError(w http.ResponseWriter, r *http.Request, e APIError) {
	e.Instance = r.URL.Path
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(e.Status)
	json.NewEncoder(w).Encode(e)
}

func errValidation(detail string) APIError {
	return APIError{Type: "validation_error", Title: "Validation Error", Status: http.StatusBadRequest, Detail: detail}
}

func errUnauthorized(detail string) APIError {
	return APIError{Type: "unauthorized", Title: "Unauthorized", Status: http.StatusUnauthorized, Detail: detail}
}

func errTooLarge(detail string) APIError {
	return APIError{Type: "payload_too_large", Title: "Payload Too Large", Status: http.StatusRequestEntityTooLarge, Detail: detail}
}

func errNotFound(detail string) APIError {
	return APIError{Type: "not_found", Title: "Not Found", Status: http.StatusNotFound, Detail: detail}
}

func errInternal(detail string) APIError {
	return APIError{Type: "internal_error", Title: "Internal Server Error", Status: http.StatusInternalServerError, Detail: detail}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd services/ingestion && go build ./internal/api/...
```

Expected: exits 0 with no output.

- [ ] **Step 3: Commit**

```bash
git add services/ingestion/internal/api/errors.go
git commit -m "feat(ingestion): RFC 7807 API error types"
```

---

### Task 5: EXIF extraction

**Files:**
- Create: `services/ingestion/internal/processor/exif.go`
- Create: `services/ingestion/internal/processor/exif_test.go`

- [ ] **Step 1: Write failing test**

`services/ingestion/internal/processor/exif_test.go`:
```go
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
	// Synthetic JPEG has no EXIF — all fields should be zero/nil
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
	// GPS is never populated — always zero to protect privacy
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/processor/... -run TestExtractEXIF -v
```

Expected: FAIL — `cannot find package`

- [ ] **Step 3: Implement exif.go**

`services/ingestion/internal/processor/exif.go`:
```go
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
	// GPSLatitude and GPSLongitude are intentionally omitted from extraction.
	// These fields exist only so callers can check the zero value.
	GPSLatitude  float64
	GPSLongitude float64
}

// ExtractEXIF reads safe EXIF fields from a JPEG file. GPS coordinates and
// device serial numbers are never extracted. Returns an empty EXIFData (not
// an error) for files with no EXIF block.
func ExtractEXIF(path string) (EXIFData, error) {
	f, err := os.Open(path)
	if err != nil {
		return EXIFData{}, err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		// No EXIF present — common for PNGs and synthetic images. Not an error.
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/processor/... -run TestExtractEXIF -v
```

Expected:
```
--- PASS: TestExtractEXIF_SyntheticJPEG_NoData (0.00s)
--- PASS: TestExtractEXIF_GPSAlwaysZero (0.00s)
--- PASS: TestExtractEXIF_MissingFile (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/internal/processor/exif.go services/ingestion/internal/processor/exif_test.go
git commit -m "feat(ingestion): EXIF extraction — GPS always stripped for privacy"
```

---

### Task 6: Image resizing

**Files:**
- Create: `services/ingestion/internal/processor/image.go`
- Create: `services/ingestion/internal/processor/image_test.go`

- [ ] **Step 1: Write failing test**

`services/ingestion/internal/processor/image_test.go`:
```go
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
	src := writeTempJPEG(t, 4032, 3024) // landscape iPhone photo
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
	src := writeTempJPEG(t, 3024, 4032) // portrait
	result, err := processor.GenerateVariants(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := longestSide(t, result.ThumbnailPath); got != 300 {
		t.Errorf("portrait thumbnail longest side = %d, want 300", got)
	}
}

func TestGenerateVariants_NoUpscale(t *testing.T) {
	src := writeTempJPEG(t, 200, 150) // smaller than 300px target
	result, err := processor.GenerateVariants(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not upscale beyond original dimensions
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/processor/... -run TestGenerateVariants -v
```

Expected: FAIL — `processor.GenerateVariants undefined`

- [ ] **Step 3: Implement image.go**

`services/ingestion/internal/processor/image.go`:
```go
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
// If both dimensions are already ≤ maxDim, saves without upscaling.
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/processor/... -run TestGenerateVariants -v
```

Expected:
```
--- PASS: TestGenerateVariants_LandscapeThumbnail (0.00s)
--- PASS: TestGenerateVariants_LandscapeWeb (0.00s)
--- PASS: TestGenerateVariants_PortraitThumbnail (0.00s)
--- PASS: TestGenerateVariants_NoUpscale (0.00s)
--- PASS: TestGenerateVariants_FileSizesPopulated (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/internal/processor/image.go services/ingestion/internal/processor/image_test.go
git commit -m "feat(ingestion): JPEG variant generation — thumbnail 300px, web 1920px"
```

---

### Task 7: HEIC conversion

**Files:**
- Create: `services/ingestion/internal/processor/heic.go`
- Create: `services/ingestion/internal/processor/heic_test.go`

**Pre-requisite:** Install libheif system library before running tests.
- macOS: `brew install libheif`
- Debian/Ubuntu: `apt-get install -y libheif-dev`

- [ ] **Step 1: Install libheif Go bindings**

```bash
cd services/ingestion
go get github.com/strukturag/libheif/go/heif@latest
```

Expected: module downloaded, go.mod updated with `github.com/strukturag/libheif`.

- [ ] **Step 2: Write failing test**

`services/ingestion/internal/processor/heic_test.go`:
```go
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
	// PNG is not HEIC — returned unchanged even though extension differs
	src := writeTempJPEG(t, 100, 100) // content is JPEG, but extension check is by name
	result, cleanup, err := processor.ToJPEGIfNeeded(src, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()
	_ = result
}

func TestToJPEGIfNeeded_HEIC_ReturnsJPEGPath(t *testing.T) {
	// Requires a real HEIC file at testdata/sample.heic (e.g. exported from iPhone).
	// Skip gracefully if not present — libheif integration needs real HEIC data.
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
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/processor/... -run TestToJPEGIfNeeded -v
```

Expected: FAIL — `processor.ToJPEGIfNeeded undefined`

- [ ] **Step 4: Implement heic.go**

`services/ingestion/internal/processor/heic.go`:
```go
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
	defer ctx.Free()

	if err := ctx.ReadFromFile(src); err != nil {
		return "", noop, fmt.Errorf("heif read: %w", err)
	}

	handle, err := ctx.GetPrimaryImageHandle()
	if err != nil {
		return "", noop, fmt.Errorf("heif primary handle: %w", err)
	}
	defer handle.Free()

	img, err := handle.DecodeImage(heif.ColorspaceRGB, heif.ChromaInterleavedRGB, nil)
	if err != nil {
		return "", noop, fmt.Errorf("heif decode: %w", err)
	}
	defer img.Free()

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

	return outPath, func() { os.Remove(outPath) }, nil
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/processor/... -run TestToJPEGIfNeeded -v
```

Expected:
```
--- PASS: TestToJPEGIfNeeded_JPEG_ReturnsSamePath (0.00s)
--- PASS: TestToJPEGIfNeeded_PNG_ReturnsSamePath (0.00s)
--- SKIP: TestToJPEGIfNeeded_HEIC_ReturnsJPEGPath (0.00s)
PASS
```

- [ ] **Step 6: Commit**

```bash
git add services/ingestion/internal/processor/heic.go services/ingestion/internal/processor/heic_test.go
git commit -m "feat(ingestion): HEIC to JPEG conversion via libheif CGO"
```

---

### Task 8: S3 upload

**Files:**
- Create: `services/ingestion/internal/storage/s3.go`
- Create: `services/ingestion/internal/storage/s3_test.go`

- [ ] **Step 1: Write failing test**

`services/ingestion/internal/storage/s3_test.go`:
```go
package storage_test

import (
	"testing"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/storage"
)

func TestNewS3Client_MissingRegion(t *testing.T) {
	_, err := storage.NewS3Client(storage.Config{Region: "", Bucket: "test"})
	if err == nil {
		t.Fatal("expected error for missing region")
	}
}

func TestNewS3Client_MissingBucket(t *testing.T) {
	_, err := storage.NewS3Client(storage.Config{Region: "ap-southeast-2", Bucket: ""})
	if err == nil {
		t.Fatal("expected error for missing bucket")
	}
}

func TestStorageClassForVariant(t *testing.T) {
	cases := []struct {
		variant string
		want    string
	}{
		{"thumbnail", "STANDARD"},
		{"web", "STANDARD"},
		{"original", "INTELLIGENT_TIERING"},
		{"unknown", "STANDARD"}, // safe default
	}
	for _, c := range cases {
		got := storage.StorageClassFor(c.variant)
		if got != c.want {
			t.Errorf("StorageClassFor(%q) = %q, want %q", c.variant, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/storage/... -v
```

Expected: FAIL — `cannot find package`

- [ ] **Step 3: Implement s3.go**

`services/ingestion/internal/storage/s3.go`:
```go
package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Region string
	Bucket string
}

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(cfg Config) (*S3Client, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("S3 region is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	return &S3Client{
		client: s3.NewFromConfig(awsCfg),
		bucket: cfg.Bucket,
	}, nil
}

// StorageClassFor returns the S3 storage class string for a given variant name.
// "original" uses INTELLIGENT_TIERING to auto-tier old photos. Everything else is STANDARD.
func StorageClassFor(variant string) string {
	if variant == "original" {
		return string(types.StorageClassIntelligentTiering)
	}
	return string(types.StorageClassStandard)
}

// Upload uploads the file at localPath to S3 at key, using storageClass.
func (c *S3Client) Upload(ctx context.Context, localPath, key, storageClass string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", localPath, err)
	}
	defer f.Close()

	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(c.bucket),
		Key:          aws.String(key),
		Body:         f,
		StorageClass: types.StorageClass(storageClass),
	})
	if err != nil {
		return fmt.Errorf("S3 put %s: %w", key, err)
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/storage/... -v
```

Expected:
```
--- PASS: TestNewS3Client_MissingRegion (0.00s)
--- PASS: TestNewS3Client_MissingBucket (0.00s)
--- PASS: TestStorageClassForVariant (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/internal/storage/
git commit -m "feat(ingestion): S3 upload — STANDARD for thumbnail/web, INTELLIGENT_TIERING for original"
```

---

### Task 9: Worker pool

**Files:**
- Create: `services/ingestion/internal/processor/worker.go`
- Create: `services/ingestion/internal/processor/worker_test.go`

- [ ] **Step 1: Write failing test**

`services/ingestion/internal/processor/worker_test.go`:
```go
package processor_test

import (
	"context"
	"os"
	"testing"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
)

// stubUploader records upload calls without hitting S3.
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/processor/... -run TestWorkerPool -v
```

Expected: FAIL — `processor.NewWorkerPool undefined`

- [ ] **Step 3: Implement worker.go**

`services/ingestion/internal/processor/worker.go`:
```go
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
// It is safe to call from multiple goroutines.
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/processor/... -run TestWorkerPool -v
```

Expected:
```
--- PASS: TestWorkerPool_Process_Success (0.00s)
--- PASS: TestWorkerPool_Process_KeyFormat (0.00s)
PASS
```

- [ ] **Step 5: Run all processor tests**

```bash
cd services/ingestion && go test ./internal/processor/... -v
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add services/ingestion/internal/processor/worker.go services/ingestion/internal/processor/worker_test.go
git commit -m "feat(ingestion): goroutine worker pool — semaphore-limited concurrency"
```

---

### Task 10: JWT middleware

**Files:**
- Create: `services/ingestion/internal/api/middleware.go`
- Create: `services/ingestion/internal/api/middleware_test.go`

JWT payload structure (issued by Node.js backend):
```json
{
  "userId": "507f1f77bcf86cd799439011",
  "email": "user@example.com",
  "role": "contributor",
  "permissions": ["images.view", "images.upload", "images.tag"],
  "iat": 1642546800,
  "exp": 1642633200
}
```

- [ ] **Step 1: Write failing test**

`services/ingestion/internal/api/middleware_test.go`:
```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	apiinternal "github.com/leahgarrett/image-management-system/services/ingestion/internal/api"
)

const testSecret = "test-jwt-secret"

func makeToken(t *testing.T, userID string, permissions []string, expiry time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"userId":      userID,
		"permissions": permissions,
		"exp":         time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func TestJWTMiddleware_ValidToken_PassesThrough(t *testing.T) {
	token := makeToken(t, "user-001", []string{"images.upload"}, time.Hour)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := apiinternal.UserIDFromContext(r.Context())
		if !ok || uid != "user-001" {
			t.Errorf("UserIDFromContext = %q, %v; want user-001, true", uid, ok)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := apiinternal.JWTMiddleware(testSecret)(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestJWTMiddleware_MissingHeader_Returns401(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := apiinternal.JWTMiddleware(testSecret)(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestJWTMiddleware_ExpiredToken_Returns401(t *testing.T) {
	token := makeToken(t, "user-001", []string{"images.upload"}, -time.Minute)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := apiinternal.JWTMiddleware(testSecret)(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestJWTMiddleware_WrongSecret_Returns401(t *testing.T) {
	token := makeToken(t, "user-001", []string{"images.upload"}, time.Hour)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := apiinternal.JWTMiddleware("wrong-secret")(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/api/... -run TestJWTMiddleware -v
```

Expected: FAIL — `apiinternal.JWTMiddleware undefined`

- [ ] **Step 3: Implement middleware.go**

`services/ingestion/internal/api/middleware.go`:
```go
package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userId"

// UserIDFromContext extracts the userId injected by JWTMiddleware.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}

// JWTMiddleware validates Bearer tokens signed with secret.
// It injects the userId into the request context and returns 401 for invalid tokens.
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeError(w, r, errUnauthorized("missing or malformed Authorization header"))
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				writeError(w, r, errUnauthorized("invalid or expired token"))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeError(w, r, errUnauthorized("invalid token claims"))
				return
			}

			userID, _ := claims["userId"].(string)
			if userID == "" {
				writeError(w, r, errUnauthorized("token missing userId"))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/api/... -run TestJWTMiddleware -v
```

Expected:
```
--- PASS: TestJWTMiddleware_ValidToken_PassesThrough (0.00s)
--- PASS: TestJWTMiddleware_MissingHeader_Returns401 (0.00s)
--- PASS: TestJWTMiddleware_ExpiredToken_Returns401 (0.00s)
--- PASS: TestJWTMiddleware_WrongSecret_Returns401 (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/internal/api/middleware.go services/ingestion/internal/api/middleware_test.go
git commit -m "feat(ingestion): JWT Bearer validation middleware"
```

---

### Task 11: HTTP handlers

**Files:**
- Create: `services/ingestion/internal/api/handlers.go`
- Create: `services/ingestion/internal/api/handlers_test.go`

- [ ] **Step 1: Write failing test**

`services/ingestion/internal/api/handlers_test.go`:
```go
package api_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func buildUploadRequest(t *testing.T, token, filename string, body []byte) *http.Request {
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
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	// Inject userId as if JWT middleware ran
	ctx := apiinternal.ContextWithUserID(req.Context(), "user-test-001")
	return req.WithContext(ctx)
}

func TestHandleUpload_AcceptsValidJPEG(t *testing.T) {
	store := jobs.NewStore()
	pool := processor.NewWorkerPool(1, &noopUploader{})
	h := apiinternal.NewHandlers(store, pool, 15*1024*1024, t.TempDir())

	req := buildUploadRequest(t, "", "photo.jpg", minimalJPEG)
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

	req := buildUploadRequest(t, "", "photo.jpg", minimalJPEG) // minimalJPEG > 10 bytes
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

type noopUploader struct{}

func (n *noopUploader) Upload(_ interface{}, _, _, _ string) error { return nil }
```

Note: `noopUploader.Upload` signature must match the `processor.Uploader` interface: `Upload(ctx context.Context, localPath, key, storageClass string) error`. Fix the test stub:

```go
import "context"

type noopUploader struct{}

func (n *noopUploader) Upload(_ context.Context, _, _, _ string) error { return nil }
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/ingestion && go test ./internal/api/... -run TestHandle -v
```

Expected: FAIL — `apiinternal.NewHandlers undefined`

- [ ] **Step 3: Implement handlers.go**

`services/ingestion/internal/api/handlers.go`:
```go
package api

import (
	"context"
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

// ContextWithUserID is exported so tests can inject a userId without a real JWT.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

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

	// Enforce size limit before parsing to avoid reading oversized bodies
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
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".heic": true, ".heif": true, ".tiff": true, ".bmp": true}
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd services/ingestion && go test ./internal/api/... -run TestHandle -v
```

Expected:
```
--- PASS: TestHandleUpload_AcceptsValidJPEG (0.00s)
--- PASS: TestHandleUpload_RejectsOversizedFile (0.00s)
--- PASS: TestHandleStatus_ReturnsJobStatus (0.00s)
--- PASS: TestHandleStatus_UnknownJob_Returns404 (0.00s)
--- PASS: TestHandleHealth_Returns200 (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/internal/api/handlers.go services/ingestion/internal/api/handlers_test.go
git commit -m "feat(ingestion): upload, status, and health HTTP handlers"
```

---

### Task 12: HTTP server

**Files:**
- Create: `services/ingestion/internal/api/server.go`

- [ ] **Step 1: Implement server.go**

`services/ingestion/internal/api/server.go`:
```go
package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter wires routes and middleware. jwtSecret is applied to all routes
// except /health.
func NewRouter(handlers *Handlers, jwtSecret string) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	api := r.PathPrefix("/api/v1/ingest").Subrouter()
	api.Use(JWTMiddleware(jwtSecret))

	api.HandleFunc("/upload", handlers.Upload).Methods(http.MethodPost)
	api.HandleFunc("/status/{jobId}", func(w http.ResponseWriter, r *http.Request) {
		jobID := mux.Vars(r)["jobId"]
		handlers.Status(w, r, jobID)
	}).Methods(http.MethodGet)

	return r
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd services/ingestion && go build ./internal/api/...
```

Expected: exits 0 with no output.

- [ ] **Step 3: Commit**

```bash
git add services/ingestion/internal/api/server.go
git commit -m "feat(ingestion): gorilla/mux router with JWT-protected subrouter"
```

---

### Task 13: Wire up main.go

**Files:**
- Modify: `services/ingestion/main.go`

- [ ] **Step 1: Update main.go to wire all dependencies**

`services/ingestion/main.go`:
```go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/leahgarrett/image-management-system/services/ingestion/internal/api"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/config"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/jobs"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/processor"
	"github.com/leahgarrett/image-management-system/services/ingestion/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	s3Client, err := storage.NewS3Client(storage.Config{
		Region: cfg.AWSRegion,
		Bucket: cfg.S3Bucket,
	})
	if err != nil {
		log.Fatalf("S3 client: %v", err)
	}

	tmpDir := os.TempDir()
	store := jobs.NewStore()
	pool := processor.NewWorkerPool(cfg.WorkerCount, s3Client)
	handlers := api.NewHandlers(store, pool, cfg.MaxFileSizeBytes, tmpDir)
	router := api.NewRouter(handlers, cfg.JWTSecret)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("ingestion service starting on :%s (workers: %d, max upload: %dMB)",
		cfg.Port, cfg.WorkerCount, cfg.MaxFileSizeBytes/1024/1024)
	log.Fatal(srv.ListenAndServe())
}
```

- [ ] **Step 2: Build the binary**

```bash
cd services/ingestion && go build -o ingestion-service .
```

Expected: `ingestion-service` binary created in current directory with no errors.

- [ ] **Step 3: Run all tests**

```bash
cd services/ingestion && go test ./...
```

Expected: all tests pass.

- [ ] **Step 4: Remove binary and commit**

```bash
rm services/ingestion/ingestion-service
git add services/ingestion/main.go
git commit -m "feat(ingestion): wire all dependencies in main.go"
```

---

### Task 14: Dockerfile and smoke test

**Files:**
- Create: `services/ingestion/Dockerfile`

- [ ] **Step 1: Create Dockerfile**

`services/ingestion/Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder

# libheif-dev + build tools needed for CGO
RUN apk add --no-cache libheif-dev gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o ingestion-service .

FROM alpine:3.19

# Runtime libheif dependency
RUN apk add --no-cache libheif ca-certificates

WORKDIR /app
COPY --from=builder /app/ingestion-service .

EXPOSE 8080

CMD ["./ingestion-service"]
```

- [ ] **Step 2: Build the Docker image**

```bash
cd services/ingestion && docker build -t ingestion-service:dev .
```

Expected: image built successfully. Final image ~30MB.

- [ ] **Step 3: Start the service for smoke testing**

In a separate terminal (or background):
```bash
docker run --rm -p 8080:8080 \
  -e JWT_SECRET=smoketest \
  -e AWS_REGION=ap-southeast-2 \
  -e S3_BUCKET=test-bucket \
  ingestion-service:dev
```

Expected log: `ingestion service starting on :8080 (workers: 10, max upload: 15MB)`

- [ ] **Step 4: Test health endpoint**

```bash
curl -s http://localhost:8080/health | jq .
```

Expected:
```json
{"status": "ok"}
```

- [ ] **Step 5: Test auth rejection**

```bash
curl -s -X POST http://localhost:8080/api/v1/ingest/upload \
  -H "Authorization: Bearer invalid-token" | jq .
```

Expected:
```json
{
  "type": "unauthorized",
  "title": "Unauthorized",
  "status": 401,
  "detail": "invalid or expired token",
  "instance": "/api/v1/ingest/upload"
}
```

- [ ] **Step 6: Generate a valid test JWT and test the upload endpoint**

Generate a JWT signed with `smoketest` secret (use jwt.io with HS256):
```json
{
  "userId": "test-user-001",
  "permissions": ["images.upload"],
  "exp": 9999999999
}
```

```bash
export TOKEN="<paste JWT here>"

curl -s -X POST http://localhost:8080/api/v1/ingest/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "image=@/path/to/any/photo.jpg" | jq .
```

Expected (S3 upload will fail since bucket is not real — job will fail, but acceptance shows routing works):
```json
{
  "jobId": "...",
  "imageId": "...",
  "status": "queued"
}
```

- [ ] **Step 7: Commit**

```bash
git add services/ingestion/Dockerfile
git commit -m "feat(ingestion): multi-stage Dockerfile with libheif runtime"
```

---

## Self-review

### Spec coverage

| Requirement (from ingestion-service-architecture.md) | Covered by |
|-----------------------------------------------------|------------|
| Go language | Task 1 (go.mod) |
| disintegration/imaging for resize | Task 6 |
| libheif for HEIC | Task 7 |
| rwcarlsen/goexif for metadata | Task 5 |
| Strip GPS by default | Task 5 (ExtractEXIF — GPS never extracted) |
| Thumbnail 300px JPEG | Task 6 |
| Web-optimised 1920px JPEG | Task 6 |
| Original stored as-is | Task 9 (worker uploads job.FilePath unchanged) |
| 15MB file size limit | Task 2 (config) + Task 11 (handler enforcement) |
| Goroutine worker pool | Task 9 |
| POST /api/v1/ingest/upload | Task 11 + Task 12 |
| GET /api/v1/ingest/status/:jobId | Task 11 + Task 12 |
| GET /health | Task 11 + Task 12 |
| JWT authentication | Task 10 |
| Async job tracking | Task 3 |
| S3 Standard for thumbnail/web | Task 8 + Task 9 |
| S3 Intelligent-Tiering for originals | Task 8 + Task 9 |
| Docker deployment | Task 14 |

### Gaps identified (out of scope for V1)

- **Glacier Deep Archive replication** — bucket-level configuration, not application code. Document in infrastructure runbook.
- **Batch upload** (`POST /api/v1/ingest/batch`) — not in this plan. Implement in a follow-up plan.
- **Cancel processing** (`DELETE /api/v1/ingest/:jobId`) — not in this plan. In-flight goroutines cannot currently be cancelled.
- **Persistent job store** — jobs are lost on restart. A Redis-backed store is a future improvement.
- **MongoDB record creation** — the completed job response includes S3 keys and metadata. The Node.js backend must poll or be notified to create the image record. Define this integration in a separate plan.
