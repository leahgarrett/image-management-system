package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leahgarrett/image-management-system/services/backend/internal/api"
	"github.com/leahgarrett/image-management-system/services/backend/internal/db"
)

func TestListUsers_AdminOnly_Returns200(t *testing.T) {
	q := &mockQuerier{
		listUsersFn: func(ctx context.Context) ([]db.User, error) {
			return []db.User{
				{Email: "admin@example.com", Role: "admin", Status: "active"},
				{Email: "user@example.com", Role: "contributor", Status: "active"},
			}, nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	ctx = api.ContextWithRole(ctx, "admin")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.ListUsers(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Users []interface{} `json:"users"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(resp.Users))
	}
}

func TestListUsers_NonAdmin_Returns403(t *testing.T) {
	handler := api.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	ctx = api.ContextWithRole(ctx, "contributor")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestInviteUser_Returns201(t *testing.T) {
	q := &mockQuerier{
		createUserFn: func(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{Email: arg.Email, Role: arg.Role, Status: arg.Status}, nil
		},
	}
	h := newTestHandlers(q)
	body := `{"email": "newuser@example.com", "name": "New User", "role": "contributor"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/invite", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	ctx = api.ContextWithRole(ctx, "admin")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.InviteUser(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateRole_Returns200(t *testing.T) {
	userUUID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	q := &mockQuerier{
		updateUserRoleFn: func(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
			return db.User{ID: userUUID, Email: "user@example.com", Role: arg.Role, Status: "active"}, nil
		},
	}
	h := newTestHandlers(q)
	body := `{"role": "admin"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+userUUID.String()+"/role", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": userUUID.String()})
	ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
	ctx = api.ContextWithRole(ctx, "admin")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.UpdateUserRole(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
