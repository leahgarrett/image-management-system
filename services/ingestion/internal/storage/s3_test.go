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
		{"unknown", "STANDARD"},
	}
	for _, c := range cases {
		got := storage.StorageClassFor(c.variant)
		if got != c.want {
			t.Errorf("StorageClassFor(%q) = %q, want %q", c.variant, got, c.want)
		}
	}
}
