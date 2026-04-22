package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalClient implements processor.Uploader by copying files to a local directory.
// Intended for local development only.
type LocalClient struct {
	dir string
}

func NewLocalClient(dir string) (*LocalClient, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create local storage dir: %w", err)
	}
	return &LocalClient{dir: dir}, nil
}

func (c *LocalClient) Upload(_ context.Context, localPath, key, _ string) error {
	dst := filepath.Join(c.dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}
