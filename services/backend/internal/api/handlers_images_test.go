package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

func TestListImages_Returns200WithPagination(t *testing.T) {
	q := &mockQuerier{
		listImagesFn: func(ctx context.Context, arg db.ListImagesParams) ([]db.Image, error) {
			return []db.Image{{ImageID: "img-001"}, {ImageID: "img-002"}}, nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images?limit=10&offset=0", nil)
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.ListImages(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Data       []interface{}          `json:"data"`
		Pagination map[string]interface{} `json:"pagination"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 images, got %d", len(resp.Data))
	}
	if resp.Pagination == nil {
		t.Error("expected pagination in response")
	}
}

func TestGetImage_Returns200WithPeopleAndTags(t *testing.T) {
	imageUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	q := &mockQuerier{
		getImageByIDFn: func(ctx context.Context, id uuid.UUID) (db.Image, error) {
			return db.Image{ID: imageUUID, ImageID: "img-001"}, nil
		},
		listImagePeopleFn: func(ctx context.Context, imageID uuid.UUID) ([]db.ImagePerson, error) {
			return []db.ImagePerson{{Name: "Alice"}, {Name: "Bob"}}, nil
		},
		listImageTagsFn: func(ctx context.Context, imageID uuid.UUID) ([]db.Tag, error) {
			return []db.Tag{{Name: "vacation"}}, nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+imageUUID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": imageUUID.String()})
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.GetImage(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	people, _ := resp["people"].([]interface{})
	if len(people) != 2 {
		t.Errorf("expected 2 people, got %v", people)
	}
	tags, _ := resp["tags"].([]interface{})
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %v", tags)
	}
}

func TestGetImage_NotFound_Returns404(t *testing.T) {
	q := &mockQuerier{
		getImageByIDFn: func(ctx context.Context, id uuid.UUID) (db.Image, error) {
			return db.Image{}, sql.ErrNoRows
		},
	}
	h := newTestHandlers(q)
	id := "00000000-0000-0000-0000-000000000099"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+id, nil)
	req = mux.SetURLVars(req, map[string]string{"id": id})
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.GetImage(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateImage_UpdatesPeopleAndTags(t *testing.T) {
	imageUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	q := &mockQuerier{
		getImageByIDFn: func(ctx context.Context, id uuid.UUID) (db.Image, error) {
			return db.Image{ID: imageUUID, ImageID: "img-001"}, nil
		},
		updateImageFn: func(ctx context.Context, arg db.UpdateImageParams) (db.Image, error) {
			return db.Image{ID: imageUUID, ImageID: "img-001"}, nil
		},
		listImageTagsFn: func(ctx context.Context, imageID uuid.UUID) ([]db.Tag, error) {
			return []db.Tag{}, nil
		},
		createTagFn: func(ctx context.Context, arg db.CreateTagParams) (db.Tag, error) {
			return db.Tag{Name: arg.Name}, nil
		},
	}
	h := newTestHandlers(q)
	body := `{"people": ["Alice", "Bob"], "tags": ["vacation", "beach"]}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/images/"+imageUUID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": imageUUID.String()})
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.UpdateImage(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateImage_NotFound_Returns404(t *testing.T) {
	q := &mockQuerier{
		getImageByIDFn: func(ctx context.Context, id uuid.UUID) (db.Image, error) {
			return db.Image{}, sql.ErrNoRows
		},
	}
	h := newTestHandlers(q)
	id := "00000000-0000-0000-0000-000000000099"
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/images/"+id, bytes.NewBufferString(`{}`))
	req = mux.SetURLVars(req, map[string]string{"id": id})
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.UpdateImage(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteImage_Returns204(t *testing.T) {
	imageUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	q := &mockQuerier{
		deleteImageFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/images/"+imageUUID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": imageUUID.String()})
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.DeleteImage(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}
