package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leahgarrett/image-management-system/services/backend/internal/api"
)

const testSecret = "test-secret"

func makeToken(t *testing.T, secret string, userID, role string, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"exp":    exp.Unix(),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tok
}

func TestJWTMiddleware_ValidToken_PassesThrough(t *testing.T) {
	tok := makeToken(t, testSecret, "user-123", "contributor", time.Now().Add(time.Hour))
	handler := api.JWTMiddleware(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := api.UserIDFromContext(r.Context())
		if !ok || uid != "user-123" {
			t.Errorf("expected userId user-123, got %q ok=%v", uid, ok)
		}
		role, ok := api.RoleFromContext(r.Context())
		if !ok || role != "contributor" {
			t.Errorf("expected role contributor, got %q ok=%v", role, ok)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestJWTMiddleware_MissingHeader_Returns401(t *testing.T) {
	handler := api.JWTMiddleware(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_ExpiredToken_Returns401(t *testing.T) {
	tok := makeToken(t, testSecret, "user-123", "contributor", time.Now().Add(-time.Hour))
	handler := api.JWTMiddleware(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_WrongSecret_Returns401(t *testing.T) {
	tok := makeToken(t, "wrong-secret", "user-123", "contributor", time.Now().Add(time.Hour))
	handler := api.JWTMiddleware(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAdmin_NonAdmin_Returns403(t *testing.T) {
	handler := api.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := api.ContextWithUserID(req.Context(), "user-123")
	ctx = api.ContextWithRole(ctx, "contributor")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRequireAdmin_Admin_PassesThrough(t *testing.T) {
	handler := api.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := api.ContextWithUserID(req.Context(), "user-123")
	ctx = api.ContextWithRole(ctx, "admin")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
