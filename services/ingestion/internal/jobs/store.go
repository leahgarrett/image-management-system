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
