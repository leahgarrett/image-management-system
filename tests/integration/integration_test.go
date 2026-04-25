//go:build integration

// Package integration_test exercises the ingestion and backend services end-to-end
// using real image files.
//
// Prerequisites:
//   - PostgreSQL database (set TEST_DATABASE_URL)
//   - Go toolchain on PATH
//
// Run with:
//
//	TEST_DATABASE_URL="postgres://user:pass@localhost:5432/testdb?sslmode=disable" \
//	  go test -v -tags integration -timeout 120s ./tests/integration/
package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	jwtSecret     = "test-integration-secret"
	ingestionPort = "18082"
	backendPort   = "18081"
	ingestionBase = "http://localhost:18082"
	backendBase   = "http://localhost:18081"
)

var (
	testDBURL     string
	authHeader    string
	testUserID    string
	testUserEmail = "integration-test@example.com"
	imageDir      string

	ingestionCmd *exec.Cmd
	backendCmd   *exec.Cmd
	storageDir   string
)

func TestMain(m *testing.M) {
	testDBURL = os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		fmt.Println("Skipping integration tests: TEST_DATABASE_URL not set.")
		fmt.Println("Run with: TEST_DATABASE_URL=postgres://... go test -tags integration ./tests/integration/")
		os.Exit(0)
	}

	repoRoot, _ := filepath.Abs("../..")
	ingestionDir := filepath.Join(repoRoot, "services", "ingestion")
	backendDir := filepath.Join(repoRoot, "services", "backend")
	imageDir = filepath.Join(repoRoot, "tests", "testdata", "images")

	var err error
	storageDir, err = os.MkdirTemp("", "integration-storage-*")
	exitOnErr("create storage dir", err)
	defer os.RemoveAll(storageDir)

	testUserID = uuid.NewString()
	token, err := makeJWT(testUserID, testUserEmail, "admin", jwtSecret)
	exitOnErr("make JWT", err)
	authHeader = "Bearer " + token

	// Start ingestion service.
	ingestionCmd = startService(ingestionDir, []string{
		"JWT_SECRET=" + jwtSecret,
		"STORAGE_BACKEND=local",
		"LOCAL_STORAGE_DIR=" + storageDir,
		"PORT=" + ingestionPort,
		"WORKER_COUNT=2",
	})
	defer ingestionCmd.Process.Kill()
	exitOnErr("ingestion health", waitForHealth(ingestionBase+"/health", 60*time.Second))

	// Start backend service — migrations run automatically on startup.
	backendCmd = startService(backendDir, []string{
		"JWT_SECRET=" + jwtSecret,
		"DATABASE_URL=" + testDBURL,
		"APP_URL=http://localhost:" + backendPort,
		"DEV_MODE=true",
		"PORT=" + backendPort,
	})
	defer backendCmd.Process.Kill()
	exitOnErr("backend health", waitForHealth(backendBase+"/health", 60*time.Second))

	// Seed test user after migrations have run.
	exitOnErr("seed test user", seedTestUser(testDBURL, testUserID, testUserEmail))
	defer cleanTestUser(testDBURL, testUserID)

	os.Exit(m.Run())
}

// ── Tests ────────────────────────────────────────────────────────────────────

// TestIngestionUpload verifies that every test image can be uploaded to the
// ingestion service and processed to completion, producing all three variants.
func TestIngestionUpload(t *testing.T) {
	images, err := filepath.Glob(filepath.Join(imageDir, "*.webp"))
	if err != nil || len(images) == 0 {
		t.Fatalf("no test images found in %s", imageDir)
	}

	for _, img := range images {
		img := img
		t.Run(filepath.Base(img), func(t *testing.T) {
			t.Parallel()

			jobID, _ := uploadImage(t, img, ingestionBase, authHeader)
			status := pollUntilComplete(t, jobID, ingestionBase, authHeader, 60*time.Second)

			keys, ok := status["keys"].(map[string]any)
			if !ok {
				t.Fatal("completed job missing keys")
			}
			for _, variant := range []string{"thumbnail", "web", "original"} {
				key, _ := keys[variant].(string)
				if key == "" {
					t.Errorf("empty key for variant %q", variant)
					continue
				}
				// Verify the file exists in local storage.
				path := filepath.Join(storageDir, filepath.FromSlash(key))
				if _, err := os.Stat(path); err != nil {
					t.Errorf("storage file missing for %q (%s): %v", variant, key, err)
				}
			}
		})
	}
}

// TestFullUploadFlow exercises the complete path:
// upload → poll → register → get → update metadata → tag suggestions → delete.
func TestFullUploadFlow(t *testing.T) {
	imagePath := filepath.Join(imageDir, "IMG_1998.webp")

	// 1. Upload to ingestion service.
	jobID, imageID := uploadImage(t, imagePath, ingestionBase, authHeader)

	// 2. Wait for processing to finish.
	status := pollUntilComplete(t, jobID, ingestionBase, authHeader, 60*time.Second)

	// 3. Register with the backend API.
	backendID := registerImage(t, imageID, filepath.Base(imagePath), status, backendBase, authHeader)
	t.Cleanup(func() { deleteImage(t, backendID, backendBase, authHeader) })

	// 4. Verify the registered image fields.
	img := getImage(t, backendID, backendBase, authHeader)
	if img["imageId"] != imageID {
		t.Errorf("imageId: got %v, want %s", img["imageId"], imageID)
	}
	if img["originalFilename"] != "IMG_1998.webp" {
		t.Errorf("originalFilename: got %v", img["originalFilename"])
	}
	if img["thumbnailKey"] == "" || img["webKey"] == "" || img["originalKey"] == "" {
		t.Errorf("storage keys missing: thumbnail=%v web=%v original=%v",
			img["thumbnailKey"], img["webKey"], img["originalKey"])
	}

	// 5. Update tags and people.
	updated := patch(t, backendID, map[string]any{
		"tags":   []string{"vacation", "beach"},
		"people": []string{"Alice", "Bob"},
	}, backendBase, authHeader)

	tags := toStringSlice(updated["tags"])
	people := toStringSlice(updated["people"])
	if !containsString(tags, "vacation") || !containsString(tags, "beach") {
		t.Errorf("tags after update: got %v, want [vacation beach]", tags)
	}
	if !containsString(people, "Alice") || !containsString(people, "Bob") {
		t.Errorf("people after update: got %v, want [Alice Bob]", people)
	}

	// 6. Confirm updates persist across a fresh GET.
	img = getImage(t, backendID, backendBase, authHeader)
	if !containsString(toStringSlice(img["tags"]), "vacation") {
		t.Errorf("tag 'vacation' missing after reload: %v", img["tags"])
	}
	if !containsString(toStringSlice(img["people"]), "Alice") {
		t.Errorf("person 'Alice' missing after reload: %v", img["people"])
	}

	// 7. Tag suggestions for prefix "vac" should return "vacation".
	suggestions := getTagSuggestions(t, "vac", backendBase, authHeader)
	if !containsString(suggestions, "vacation") {
		t.Errorf("'vacation' not in suggestions for 'vac': %v", suggestions)
	}
}

// TestOccasionFiltering registers two images with different occasions and
// verifies the occasion query filter returns only the matching image.
func TestOccasionFiltering(t *testing.T) {
	testImages := []struct {
		file     string
		occasion string
	}{
		{"chatgpt-people-city-rain.webp", "birthday"},
		{"chatgpt-people-laughing.webp", "wedding"},
	}

	type entry struct {
		backendID string
		occasion  string
	}
	var registered []entry

	for _, tc := range testImages {
		jobID, imageID := uploadImage(t, filepath.Join(imageDir, tc.file), ingestionBase, authHeader)
		status := pollUntilComplete(t, jobID, ingestionBase, authHeader, 60*time.Second)
		bid := registerImage(t, imageID, tc.file, status, backendBase, authHeader)
		registered = append(registered, entry{backendID: bid, occasion: tc.occasion})
		patch(t, bid, map[string]any{"occasionCategory": tc.occasion}, backendBase, authHeader)
	}
	t.Cleanup(func() {
		for _, e := range registered {
			deleteImage(t, e.backendID, backendBase, authHeader)
		}
	})

	// Filter by birthday — only the first image should appear.
	birthdayIDs := listImageIDsByOccasion(t, "birthday", backendBase, authHeader)
	if !containsString(birthdayIDs, registered[0].backendID) {
		t.Errorf("birthday image %s not in filtered results: %v", registered[0].backendID, birthdayIDs)
	}
	if containsString(birthdayIDs, registered[1].backendID) {
		t.Errorf("wedding image %s should not appear in birthday filter", registered[1].backendID)
	}
}

// TestTagSuggestions verifies that prefix search returns matching tags only.
func TestTagSuggestions(t *testing.T) {
	imagePath := filepath.Join(imageDir, "IMG_2119.webp")

	jobID, imageID := uploadImage(t, imagePath, ingestionBase, authHeader)
	status := pollUntilComplete(t, jobID, ingestionBase, authHeader, 60*time.Second)
	backendID := registerImage(t, imageID, filepath.Base(imagePath), status, backendBase, authHeader)
	t.Cleanup(func() { deleteImage(t, backendID, backendBase, authHeader) })

	patch(t, backendID, map[string]any{
		"tags": []string{"birthday", "beach", "sunset"},
	}, backendBase, authHeader)

	suggestions := getTagSuggestions(t, "b", backendBase, authHeader)
	if !containsString(suggestions, "birthday") {
		t.Errorf("'birthday' missing from suggestions for 'b': %v", suggestions)
	}
	if !containsString(suggestions, "beach") {
		t.Errorf("'beach' missing from suggestions for 'b': %v", suggestions)
	}
	if containsString(suggestions, "sunset") {
		t.Errorf("'sunset' should not appear in suggestions for 'b': %v", suggestions)
	}
}

// TestDeleteImage verifies that a deleted image returns 404.
func TestDeleteImage(t *testing.T) {
	imagePath := filepath.Join(imageDir, "IMG_7216.webp")

	jobID, imageID := uploadImage(t, imagePath, ingestionBase, authHeader)
	status := pollUntilComplete(t, jobID, ingestionBase, authHeader, 60*time.Second)
	backendID := registerImage(t, imageID, filepath.Base(imagePath), status, backendBase, authHeader)

	deleteImage(t, backendID, backendBase, authHeader)

	// GET should now return 404.
	req, _ := http.NewRequest(http.MethodGet, backendBase+"/api/v1/images/"+backendID, nil)
	req.Header.Set("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET after delete: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", resp.StatusCode)
	}
}

// TestReplaceTagsOnUpdate verifies that PATCH replaces the tag list (not appends).
func TestReplaceTagsOnUpdate(t *testing.T) {
	imagePath := filepath.Join(imageDir, "IMG_2256.webp")

	jobID, imageID := uploadImage(t, imagePath, ingestionBase, authHeader)
	status := pollUntilComplete(t, jobID, ingestionBase, authHeader, 60*time.Second)
	backendID := registerImage(t, imageID, filepath.Base(imagePath), status, backendBase, authHeader)
	t.Cleanup(func() { deleteImage(t, backendID, backendBase, authHeader) })

	// First update: two tags.
	patch(t, backendID, map[string]any{"tags": []string{"alpha", "beta"}}, backendBase, authHeader)

	// Second update: replace with one different tag.
	updated := patch(t, backendID, map[string]any{"tags": []string{"gamma"}}, backendBase, authHeader)
	tags := toStringSlice(updated["tags"])

	if containsString(tags, "alpha") || containsString(tags, "beta") {
		t.Errorf("old tags still present after replace: %v", tags)
	}
	if !containsString(tags, "gamma") {
		t.Errorf("'gamma' missing after replace: %v", tags)
	}
}

// ── helpers used only in this file ──────────────────────────────────────────

// patch sends PATCH /api/v1/images/{id} with an arbitrary JSON body.
func patch(t *testing.T, id string, body map[string]any, base, auth string) map[string]any {
	t.Helper()
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPatch, base+"/api/v1/images/"+id, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /images/%s: %v", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("PATCH /images/%s: expected 200, got %d: %s", id, resp.StatusCode, b)
	}
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// listImageIDsByOccasion returns the backend UUIDs of images matching the occasion filter.
func listImageIDsByOccasion(t *testing.T, occasion, base, auth string) []string {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, base+"/api/v1/images?occasion="+occasion, nil)
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /images?occasion=%s: %v", occasion, err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []map[string]any `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	ids := make([]string, 0, len(result.Data))
	for _, img := range result.Data {
		if id, ok := img["id"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func exitOnErr(label string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL %s: %v\n", label, err)
		os.Exit(1)
	}
}
