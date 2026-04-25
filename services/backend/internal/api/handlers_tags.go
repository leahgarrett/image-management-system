package api

import (
	"encoding/json"
	"net/http"
)

type tagResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsageCount int32  `json:"usageCount"`
}

// ListTags handles GET /api/v1/tags.
func (h *Handlers) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.q.ListTags(r.Context())
	if err != nil {
		writeError(w, r, errInternal("failed to list tags"))
		return
	}

	resp := make([]tagResponse, 0, len(tags))
	for _, t := range tags {
		resp = append(resp, tagResponse{
			ID:         t.ID.String(),
			Name:       t.Name,
			UsageCount: t.UsageCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tags": resp})
}

// TagSuggestions handles GET /api/v1/tags/suggestions?q=xxx.
func (h *Handlers) TagSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, r, errValidation("q query parameter is required"))
		return
	}

	tags, err := h.q.SearchTags(r.Context(), query)
	if err != nil {
		writeError(w, r, errInternal("failed to search tags"))
		return
	}

	resp := make([]tagResponse, 0, len(tags))
	for _, t := range tags {
		resp = append(resp, tagResponse{
			ID:         t.ID.String(),
			Name:       t.Name,
			UsageCount: t.UsageCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tags": resp})
}
