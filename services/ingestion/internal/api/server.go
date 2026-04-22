package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter wires routes and middleware. jwtSecret is applied to all routes
// except /health.
func NewRouter(handlers *Handlers, jwtSecret string) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	api := r.PathPrefix("/api/v1/ingest").Subrouter()
	api.Use(JWTMiddleware(jwtSecret))

	api.HandleFunc("/upload", handlers.Upload).Methods(http.MethodPost)
	api.HandleFunc("/status/{jobId}", func(w http.ResponseWriter, r *http.Request) {
		jobID := mux.Vars(r)["jobId"]
		handlers.Status(w, r, jobID)
	}).Methods(http.MethodGet)

	return r
}
