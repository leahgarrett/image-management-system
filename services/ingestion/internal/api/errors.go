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
