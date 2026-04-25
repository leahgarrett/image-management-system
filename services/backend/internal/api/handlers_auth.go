package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leahgarrett/image-management-system/services/backend/internal/db"
	"github.com/leahgarrett/image-management-system/services/backend/internal/mailer"
)

// Handlers holds all handler dependencies.
type Handlers struct {
	q         db.Querier
	mailer    mailer.Mailer
	appURL    string
	jwtSecret string
	devMode   bool
}

// NewHandlers constructs a Handlers instance.
func NewHandlers(q db.Querier, m mailer.Mailer, appURL, jwtSecret string, devMode bool) *Handlers {
	return &Handlers{q: q, mailer: m, appURL: appURL, jwtSecret: jwtSecret, devMode: devMode}
}

// Health is the liveness probe endpoint.
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type loginRequest struct {
	Email string `json:"email"`
}

// Login handles POST /api/v1/auth/login.
// It finds or creates the user, generates a magic link token, and sends it via email.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeError(w, r, errValidation("email is required"))
		return
	}

	ctx := r.Context()

	user, err := h.q.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Count existing users to determine role for first user.
			users, listErr := h.q.ListUsers(ctx)
			role := "contributor"
			if listErr == nil && len(users) == 0 {
				role = "admin"
			}
			user, err = h.q.CreateUser(ctx, db.CreateUserParams{
				Email:  req.Email,
				Role:   role,
				Status: "active",
			})
			if err != nil {
				writeError(w, r, errInternal("failed to create user"))
				return
			}
		} else {
			writeError(w, r, errInternal("failed to look up user"))
			return
		}
	}

	// Generate raw token: 32 random bytes, hex-encoded.
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		writeError(w, r, errInternal("failed to generate token"))
		return
	}
	rawToken := hex.EncodeToString(rawBytes)

	// Store SHA-256 hash of the token.
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	expiresAt := time.Now().Add(15 * time.Minute)
	_, err = h.q.CreateMagicLinkToken(ctx, db.CreateMagicLinkTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		writeError(w, r, errInternal("failed to create token"))
		return
	}

	magicLinkURL := fmt.Sprintf("%s/auth/verify?token=%s", h.appURL, rawToken)
	if err := h.mailer.SendMagicLink(req.Email, magicLinkURL); err != nil {
		writeError(w, r, errInternal("failed to send magic link"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Magic link sent"})
}

// Verify handles GET /api/v1/auth/verify?token=xxx.
func (h *Handlers) Verify(w http.ResponseWriter, r *http.Request) {
	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		writeError(w, r, errValidation("token query parameter is required"))
		return
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	ctx := r.Context()
	record, err := h.q.GetMagicLinkTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, r, errUnauthorized("invalid or expired token"))
			return
		}
		writeError(w, r, errInternal("failed to look up token"))
		return
	}

	if record.UsedAt.Valid {
		writeError(w, r, errUnauthorized("token has already been used"))
		return
	}
	if time.Now().After(record.ExpiresAt) {
		writeError(w, r, errUnauthorized("token has expired"))
		return
	}

	if err := h.q.MarkTokenUsed(ctx, record.ID); err != nil {
		writeError(w, r, errInternal("failed to mark token used"))
		return
	}

	user, err := h.q.GetUserByID(ctx, record.UserID)
	if err != nil {
		writeError(w, r, errInternal("failed to look up user"))
		return
	}

	if err := h.q.UpdateUserLastLogin(ctx, user.ID); err != nil {
		writeError(w, r, errInternal("failed to update last login"))
		return
	}

	// Issue JWT.
	claims := jwt.MapClaims{
		"userId": user.ID.String(),
		"email":  user.Email,
		"role":   user.Role,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
	if err != nil {
		writeError(w, r, errInternal("failed to issue token"))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenStr,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(24 * time.Hour / time.Second),
		Secure:   !h.devMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Authenticated"})
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   0,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}

// issueJWT is a helper used by tests to issue tokens directly.
func issueJWT(secret, userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"role":   role,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

