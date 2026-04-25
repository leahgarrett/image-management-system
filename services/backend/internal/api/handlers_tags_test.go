package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leahgarrett/image-management-system/services/backend/internal/api"
	"github.com/leahgarrett/image-management-system/services/backend/internal/db"
)

// localTagResponse is a local struct for decoding tag JSON in tests.
// (tagResponse in the api package is unexported)
type localTagResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsageCount int32  `json:"usageCount"`
}

func TestListTags_Returns200(t *testing.T) {
	q := &mockQuerier{
		listTagsFn: func(ctx context.Context) ([]db.Tag, error) {
			return []db.Tag{
				{Name: "vacation", UsageCount: 5},
				{Name: "beach", UsageCount: 3},
			}, nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.ListTags(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Tags []localTagResponse `json:"tags"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(resp.Tags))
	}
}

func TestSuggestions_FiltersCorrectly(t *testing.T) {
	q := &mockQuerier{
		searchTagsFn: func(ctx context.Context, query string) ([]db.Tag, error) {
			if query != "vac" {
				return []db.Tag{}, nil
			}
			return []db.Tag{{Name: "vacation", UsageCount: 5}}, nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tags/suggestions?q=vac", nil)
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.TagSuggestions(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Tags []localTagResponse `json:"tags"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Tags) != 1 || resp.Tags[0].Name != "vacation" {
		t.Errorf("unexpected suggestions: %v", resp.Tags)
	}
}
