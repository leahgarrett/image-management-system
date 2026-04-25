//go:build integration

package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// makeJWT creates a signed HS256 token accepted by both services.
func makeJWT(userID, email, role, secret string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"role":   role,
		"exp":    time.Now().Add(2 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// startService runs "go run ." in dir with the given env vars appended to the current environment.
func startService(dir string, envVars []string) *exec.Cmd {
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), envVars...)
	cmd.Stdout = os.Stderr // show service logs under go test -v
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("startService(%s): %v", dir, err))
	}
	return cmd
}

// waitForHealth polls url until it returns 200 or timeout is exceeded.
func waitForHealth(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("service at %s did not become healthy within %v", url, timeout)
}

// seedTestUser inserts the integration test user into postgres.
// Uses ON CONFLICT so it is safe to call multiple times.
func seedTestUser(dbURL, userID, email string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()
	_, err = db.Exec(`
		INSERT INTO users (id, email, name, role, status, created_at)
		VALUES ($1, $2, 'Integration Test', 'admin', 'active', NOW())
		ON CONFLICT (id) DO NOTHING
	`, userID, email)
	return err
}

// cleanTestUser removes all images/tags owned by the test user, then the user itself.
func cleanTestUser(dbURL, userID string) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return
	}
	defer db.Close()
	db.Exec(`DELETE FROM images WHERE uploaded_by = $1`, userID)
	db.Exec(`DELETE FROM users WHERE id = $1`, userID)
}

// uploadImage POSTs imagePath as multipart to the ingestion service.
// Returns jobID and imageID from the 202 response.
func uploadImage(t *testing.T, imagePath, ingestionBase, authHeader string) (jobID, imageID string) {
	t.Helper()

	f, err := os.Open(imagePath)
	if err != nil {
		t.Fatalf("open test image %s: %v", imagePath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		t.Fatalf("copy image into multipart: %v", err)
	}
	mw.Close()

	req, _ := http.NewRequest(http.MethodPost, ingestionBase+"/api/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload: expected 202, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result["jobId"].(string), result["imageId"].(string)
}

// pollUntilComplete polls the ingestion status endpoint until the job is completed or fails.
func pollUntilComplete(t *testing.T, jobID, ingestionBase, authHeader string, timeout time.Duration) map[string]any {
	t.Helper()
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 5 * time.Second}

	for time.Now().Before(deadline) {
		req, _ := http.NewRequest(http.MethodGet, ingestionBase+"/api/v1/ingest/status/"+jobID, nil)
		req.Header.Set("Authorization", authHeader)

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		var result map[string]any
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		switch result["status"] {
		case "completed":
			return result
		case "failed":
			t.Fatalf("ingestion job %s failed: %v", jobID, result["error"])
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("job %s did not complete within %v", jobID, timeout)
	return nil
}

// registerImage calls POST /api/v1/images on the backend.
// Returns the backend image UUID.
func registerImage(t *testing.T, imageID, originalFilename string, status map[string]any, backendBase, authHeader string) string {
	t.Helper()

	keys, _ := status["keys"].(map[string]any)
	meta, _ := status["metadata"].(map[string]any)

	width, height := int32(0), int32(0)
	if w, ok := meta["width"].(float64); ok {
		width = int32(w)
	}
	if h, ok := meta["height"].(float64); ok {
		height = int32(h)
	}

	captureDate, _ := meta["captureDate"].(string)
	cameraMake, _ := meta["cameraMake"].(string)
	cameraModel, _ := meta["cameraModel"].(string)

	body := map[string]any{
		"imageId":          imageID,
		"originalFilename": originalFilename,
		"thumbnailKey":     keys["thumbnail"],
		"webKey":           keys["web"],
		"originalKey":      keys["original"],
		"width":            width,
		"height":           height,
		"metadata": map[string]any{
			"captureDate": captureDate,
			"cameraMake":  cameraMake,
			"cameraModel": cameraModel,
		},
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, backendBase+"/api/v1/images", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("register image: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("register image: expected 201, got %d: %s", resp.StatusCode, b)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result["id"].(string)
}

// deleteImage calls DELETE /api/v1/images/{id} on the backend.
func deleteImage(t *testing.T, id, backendBase, authHeader string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, backendBase+"/api/v1/images/"+id, nil)
	req.Header.Set("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Logf("deleteImage(%s): %v", id, err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Logf("deleteImage(%s): expected 204, got %d", id, resp.StatusCode)
	}
}

// patchImage calls PATCH /api/v1/images/{id} with the given tags and people.
func patchImage(t *testing.T, id string, tags, people []string, backendBase, authHeader string) map[string]any {
	t.Helper()
	body := map[string]any{"tags": tags, "people": people}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPatch, backendBase+"/api/v1/images/"+id, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("patchImage(%s): %v", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("patchImage(%s): expected 200, got %d: %s", id, resp.StatusCode, b)
	}
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// getImage calls GET /api/v1/images/{id} and returns the decoded JSON.
func getImage(t *testing.T, id, backendBase, authHeader string) map[string]any {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, backendBase+"/api/v1/images/"+id, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("getImage(%s): %v", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("getImage(%s): expected 200, got %d: %s", id, resp.StatusCode, b)
	}
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// getTagSuggestions calls GET /api/v1/tags/suggestions?q=<query>.
func getTagSuggestions(t *testing.T, query, backendBase, authHeader string) []string {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, backendBase+"/api/v1/tags/suggestions?q="+query, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("getTagSuggestions(%q): %v", query, err)
	}
	defer resp.Body.Close()

	var result struct {
		Tags []struct {
			Name string `json:"name"`
		} `json:"tags"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	names := make([]string, 0, len(result.Tags))
	for _, tg := range result.Tags {
		names = append(names, tg.Name)
	}
	return names
}

// containsString reports whether slice contains s.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// toStringSlice converts []any to []string.
func toStringSlice(v any) []string {
	raw, _ := v.([]any)
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
