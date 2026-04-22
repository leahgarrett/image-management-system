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
