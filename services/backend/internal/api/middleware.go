package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userId"
const roleKey contextKey = "role"

// UserIDFromContext extracts the userId injected by JWTMiddleware.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}

// RoleFromContext extracts the role injected by JWTMiddleware.
func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok && role != ""
}

// ContextWithUserID returns a copy of ctx with userID stored — exported for tests.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// ContextWithRole returns a copy of ctx with role stored — exported for tests.
func ContextWithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

// JWTMiddleware validates tokens from the Authorization header (Bearer) or the
// auth_token httpOnly cookie. It injects userId and role into the request context.
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := ""

			// Prefer Authorization header; fall back to cookie.
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				cookie, err := r.Cookie("auth_token")
				if err != nil {
					writeError(w, r, errUnauthorized("missing or malformed Authorization header or auth_token cookie"))
					return
				}
				tokenStr = cookie.Value
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				writeError(w, r, errUnauthorized("invalid or expired token"))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeError(w, r, errUnauthorized("invalid token claims"))
				return
			}

			userID, _ := claims["userId"].(string)
			if userID == "" {
				writeError(w, r, errUnauthorized("token missing userId"))
				return
			}

			role, _ := claims["role"].(string)

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, roleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin returns 403 if the role in context is not "admin".
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := RoleFromContext(r.Context())
		if !ok || role != "admin" {
			writeError(w, r, errForbidden("admin role required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
