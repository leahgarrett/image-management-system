package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	apiinternal "github.com/leahgarrett/image-management-system/services/ingestion/internal/api"
)

const testSecret = "test-jwt-secret"

func makeToken(t *testing.T, userID string, permissions []string, expiry time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"userId":      userID,
		"permissions": permissions,
		"exp":         time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func TestJWTMiddleware_ValidToken_PassesThrough(t *testing.T) {
	token := makeToken(t, "user-001", []string{"images.upload"}, time.Hour)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := apiinternal.UserIDFromContext(r.Context())
		if !ok || uid != "user-001" {
			t.Errorf("UserIDFromContext = %q, %v; want user-001, true", uid, ok)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := apiinternal.JWTMiddleware(testSecret)(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestJWTMiddleware_MissingHeader_Returns401(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := apiinternal.JWTMiddleware(testSecret)(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestJWTMiddleware_ExpiredToken_Returns401(t *testing.T) {
	token := makeToken(t, "user-001", []string{"images.upload"}, -time.Minute)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := apiinternal.JWTMiddleware(testSecret)(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestJWTMiddleware_WrongSecret_Returns401(t *testing.T) {
	token := makeToken(t, "user-001", []string{"images.upload"}, time.Hour)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := apiinternal.JWTMiddleware("wrong-secret")(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
