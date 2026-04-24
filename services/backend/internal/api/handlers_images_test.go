package api_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leahgarrett/image-management-system/services/backend/internal/api"
	"github.com/leahgarrett/image-management-system/services/backend/internal/db"
)

func TestRegisterImage_ValidRequest_Returns201(t *testing.T) {
	q := &mockQuerier{
		createImageFn: func(ctx context.Context, arg db.CreateImageParams) (db.Image, error) {
			return db.Image{ImageID: arg.ImageID}, nil
		},
	}
	h := newTestHandlers(q)
	body := `{
		"imageId": "img-abc-123",
		"originalFilename": "photo.jpg",
		"thumbnailKey": "user/img/thumbnail.jpg",
		"webKey": "user/img/web.jpg",
		"originalKey": "user/img/original.jpg",
		"thumbnailSize": 30000,
		"webSize": 300000,
		"originalSize": 15000000,
		"width": 4032,
		"height": 3024,
		"metadata": {"captureDate": "2024-06-15T10:30:00Z", "cameraMake": "Apple", "cameraModel": "iPhone 14 Pro"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.RegisterImage(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRegisterImage_MissingImageID_Returns400(t *testing.T) {
	h := newTestHandlers(&mockQuerier{})
	body := `{"originalFilename": "photo.jpg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.RegisterImage(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestRegisterImage_SetsUploadedByFromContext(t *testing.T) {
	const testUserID = "00000000-0000-0000-0000-000000000001"
	var capturedUploadedBy string
	q := &mockQuerier{
		createImageFn: func(ctx context.Context, arg db.CreateImageParams) (db.Image, error) {
			capturedUploadedBy = arg.UploadedBy.UUID.String()
			return db.Image{ImageID: arg.ImageID}, nil
		},
	}
	h := newTestHandlers(q)
	body := `{"imageId": "img-test-001", "metadata": {}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := api.ContextWithUserID(req.Context(), testUserID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.RegisterImage(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if capturedUploadedBy != testUserID {
		t.Errorf("expected uploadedBy %s, got %s", testUserID, capturedUploadedBy)
	}
}
