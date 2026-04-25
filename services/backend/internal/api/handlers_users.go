package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leahgarrett/image-management-system/services/backend/internal/db"
)

type userResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Name        string  `json:"name,omitempty"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"createdAt"`
	LastLoginAt *string `json:"lastLoginAt,omitempty"`
}

func userToResponse(u db.User) userResponse {
	resp := userResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name.String,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
	if u.LastLoginAt.Valid {
		s := u.LastLoginAt.Time.Format(time.RFC3339)
		resp.LastLoginAt = &s
	}
	return resp
}

// ListUsers handles GET /api/v1/users (admin only).
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.q.ListUsers(r.Context())
	if err != nil {
		writeError(w, r, errInternal("failed to list users"))
		return
	}

	resp := make([]userResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, userToResponse(u))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"users": resp})
}

type inviteUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// InviteUser handles POST /api/v1/users/invite (admin only).
func (h *Handlers) InviteUser(w http.ResponseWriter, r *http.Request) {
	var req inviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeError(w, r, errValidation("email is required"))
		return
	}
	if req.Role == "" {
		req.Role = "contributor"
	}

	inviterIDStr, _ := UserIDFromContext(r.Context())
	var invitedBy uuid.NullUUID
	if parsed, err := uuid.Parse(inviterIDStr); err == nil {
		invitedBy = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	user, err := h.q.CreateUser(r.Context(), db.CreateUserParams{
		Email:     req.Email,
		Name:      sql.NullString{String: req.Name, Valid: req.Name != ""},
		Role:      req.Role,
		Status:    "invited",
		InvitedBy: invitedBy,
	})
	if err != nil {
		writeError(w, r, errInternal("failed to create user"))
		return
	}

	// Generate magic link token (72h expiry for invitations).
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		writeError(w, r, errInternal("failed to generate token"))
		return
	}
	rawToken := hex.EncodeToString(rawBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	if _, err := h.q.CreateMagicLinkToken(r.Context(), db.CreateMagicLinkTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(72 * time.Hour),
	}); err != nil {
		writeError(w, r, errInternal("failed to create invite token"))
		return
	}

	magicLinkURL := fmt.Sprintf("%s/auth/verify?token=%s", h.appURL, rawToken)
	if err := h.mailer.SendMagicLink(req.Email, magicLinkURL); err != nil {
		writeError(w, r, errInternal("failed to send invite email"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userToResponse(user))
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

// UpdateUserRole handles PATCH /api/v1/users/{id}/role (admin only).
func (h *Handlers) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, r, errValidation("invalid user id"))
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Role == "" {
		writeError(w, r, errValidation("role is required"))
		return
	}
	if req.Role != "admin" && req.Role != "contributor" {
		writeError(w, r, errValidation("role must be admin or contributor"))
		return
	}

	user, err := h.q.UpdateUserRole(r.Context(), db.UpdateUserRoleParams{ID: id, Role: req.Role})
	if err != nil {
		writeError(w, r, errInternal("failed to update role"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userToResponse(user))
}
