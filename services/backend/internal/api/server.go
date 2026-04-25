package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter constructs the gorilla/mux router with all routes registered.
func NewRouter(h *Handlers, jwtSecret string) http.Handler {
	r := mux.NewRouter()

	// Health probe — no auth required.
	r.HandleFunc("/health", h.Health).Methods(http.MethodGet)

	// Auth routes — no JWT middleware.
	auth := r.PathPrefix("/api/v1/auth").Subrouter()
	auth.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	auth.HandleFunc("/verify", h.Verify).Methods(http.MethodGet)
	auth.HandleFunc("/logout", h.Logout).Methods(http.MethodPost)

	// Authenticated API routes.
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(JWTMiddleware(jwtSecret))

	api.HandleFunc("/images", h.ListImages).Methods(http.MethodGet)
	api.HandleFunc("/images", h.RegisterImage).Methods(http.MethodPost)
	api.HandleFunc("/images/{id}", h.GetImage).Methods(http.MethodGet)
	api.HandleFunc("/images/{id}", h.UpdateImage).Methods(http.MethodPatch)
	api.HandleFunc("/images/{id}", h.DeleteImage).Methods(http.MethodDelete)

	api.HandleFunc("/tags", h.ListTags).Methods(http.MethodGet)
	api.HandleFunc("/tags/suggestions", h.TagSuggestions).Methods(http.MethodGet)

	// Admin-only user management routes.
	admin := api.PathPrefix("/users").Subrouter()
	admin.Use(RequireAdmin)
	admin.HandleFunc("", h.ListUsers).Methods(http.MethodGet)
	admin.HandleFunc("/invite", h.InviteUser).Methods(http.MethodPost)
	admin.HandleFunc("/{id}/role", h.UpdateUserRole).Methods(http.MethodPatch)

	return r
}
