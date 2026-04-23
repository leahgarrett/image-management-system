# Backend Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go REST API that manages image metadata in PostgreSQL, authenticates users via magic link, and serves CRUD endpoints for images, tags, and users.

**Architecture:** gorilla/mux HTTP server with PostgreSQL backend (pgx + sqlc for type-safe queries). Magic link auth issues JWT tokens stored in httpOnly cookies. Migrations run automatically on startup via golang-migrate embedded in the binary.

**Tech Stack:** Go 1.21, gorilla/mux, jackc/pgx/v5, sqlc, golang-migrate, golang-jwt/jwt/v5, google/uuid, golang.org/x/crypto

---

## Task 1: Project scaffold

Create `services/backend/go.mod`, `sqlc.yaml`, and `.env.example`.

- [ ] Create `services/backend/go.mod`:

```
module github.com/leahgarrett/image-management-system/services/backend

go 1.21

require (
    github.com/golang-jwt/jwt/v5 v5.2.1
    github.com/golang-migrate/migrate/v4 v4.17.0
    github.com/google/uuid v1.6.0
    github.com/gorilla/mux v1.8.1
    github.com/jackc/pgx/v5 v5.5.5
    golang.org/x/crypto v0.21.0
)
```

- [ ] Create `services/backend/sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/db/query/"
    schema: "migrations/"
    gen:
      go:
        package: "db"
        out: "internal/db"
        emit_json_tags: true
        emit_interface: true
        emit_empty_slices: true
```

- [ ] Create `services/backend/.env.example`:

```
PORT=8081
DATABASE_URL=postgres://backend:backend@localhost:5432/imagedb?sslmode=disable
JWT_SECRET=changeme
APP_URL=http://localhost:3000
DEV_MODE=true
SMTP_HOST=
SMTP_PORT=587
SMTP_FROM=
```

- [ ] Run from `services/backend/`:

```bash
go mod download
```

Expected: no errors, `go.sum` created.

**Commit:** `feat(backend): project scaffold`

---

## Task 2: Config

Create `services/backend/internal/config/config.go` and `config_test.go`, replicating the ingestion service pattern exactly.

- [ ] Create `services/backend/internal/config/config.go`:

```go
package config

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    Port        string
    DatabaseURL string
    JWTSecret   string
    AppURL      string
    DevMode     bool
    SMTPHost    string
    SMTPPort    int
    SMTPFrom    string
}

func Load() (*Config, error) {
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }

    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        return nil, fmt.Errorf("JWT_SECRET is required")
    }

    appURL := os.Getenv("APP_URL")
    if appURL == "" {
        return nil, fmt.Errorf("APP_URL is required")
    }

    port := getEnvOrDefault("PORT", "8081")

    devMode := false
    if v := os.Getenv("DEV_MODE"); v == "true" {
        devMode = true
    }

    smtpHost := os.Getenv("SMTP_HOST")
    smtpFrom := os.Getenv("SMTP_FROM")

    if !devMode {
        if smtpHost == "" {
            return nil, fmt.Errorf("SMTP_HOST is required when DEV_MODE is not true")
        }
        if smtpFrom == "" {
            return nil, fmt.Errorf("SMTP_FROM is required when DEV_MODE is not true")
        }
    }

    smtpPort := 587
    if v := os.Getenv("SMTP_PORT"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 1 {
            return nil, fmt.Errorf("SMTP_PORT must be a positive integer")
        }
        smtpPort = n
    }

    return &Config{
        Port:        port,
        DatabaseURL: databaseURL,
        JWTSecret:   jwtSecret,
        AppURL:      appURL,
        DevMode:     devMode,
        SMTPHost:    smtpHost,
        SMTPPort:    smtpPort,
        SMTPFrom:    smtpFrom,
    }, nil
}

func getEnvOrDefault(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
```

- [ ] Create `services/backend/internal/config/config_test.go`:

```go
package config_test

import (
    "testing"

    "github.com/leahgarrett/image-management-system/services/backend/internal/config"
)

func setBase(t *testing.T) {
    t.Helper()
    t.Setenv("DATABASE_URL", "postgres://backend:backend@localhost:5432/imagedb?sslmode=disable")
    t.Setenv("JWT_SECRET", "testsecret")
    t.Setenv("APP_URL", "http://localhost:3000")
    t.Setenv("DEV_MODE", "true")
}

func TestLoad_Defaults(t *testing.T) {
    setBase(t)
    cfg, err := config.Load()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if cfg.Port != "8081" {
        t.Errorf("expected port 8081, got %s", cfg.Port)
    }
    if cfg.SMTPPort != 587 {
        t.Errorf("expected SMTP port 587, got %d", cfg.SMTPPort)
    }
    if cfg.DevMode != true {
        t.Error("expected DevMode true")
    }
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
    t.Setenv("DATABASE_URL", "")
    t.Setenv("JWT_SECRET", "testsecret")
    t.Setenv("APP_URL", "http://localhost:3000")
    t.Setenv("DEV_MODE", "true")
    _, err := config.Load()
    if err == nil {
        t.Fatal("expected error for missing DATABASE_URL")
    }
}

func TestLoad_MissingJWTSecret(t *testing.T) {
    t.Setenv("DATABASE_URL", "postgres://backend:backend@localhost:5432/imagedb?sslmode=disable")
    t.Setenv("JWT_SECRET", "")
    t.Setenv("APP_URL", "http://localhost:3000")
    t.Setenv("DEV_MODE", "true")
    _, err := config.Load()
    if err == nil {
        t.Fatal("expected error for missing JWT_SECRET")
    }
}

func TestLoad_SMTPRequiredWhenNotDevMode(t *testing.T) {
    t.Setenv("DATABASE_URL", "postgres://backend:backend@localhost:5432/imagedb?sslmode=disable")
    t.Setenv("JWT_SECRET", "testsecret")
    t.Setenv("APP_URL", "http://localhost:3000")
    t.Setenv("DEV_MODE", "false")
    t.Setenv("SMTP_HOST", "")
    t.Setenv("SMTP_FROM", "")
    _, err := config.Load()
    if err == nil {
        t.Fatal("expected error for missing SMTP config when DEV_MODE=false")
    }
}

func TestLoad_SMTPNotRequiredInDevMode(t *testing.T) {
    setBase(t)
    t.Setenv("SMTP_HOST", "")
    t.Setenv("SMTP_FROM", "")
    _, err := config.Load()
    if err != nil {
        t.Fatalf("unexpected error in dev mode without SMTP: %v", err)
    }
}
```

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/config/...
```

Expected output:
```
ok      github.com/leahgarrett/image-management-system/services/backend/internal/config
```

**Commit:** `feat(backend): config loading`

---

## Task 3: Database migrations

Create all 6 up/down migration pairs in `services/backend/migrations/`.

- [ ] Create `migrations/001_create_users.up.sql`:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT NOT NULL UNIQUE,
  name          TEXT,
  role          TEXT NOT NULL DEFAULT 'contributor'
    CHECK (role IN ('admin', 'contributor')),
  status        TEXT NOT NULL DEFAULT 'invited'
    CHECK (status IN ('invited', 'active', 'suspended')),
  invited_by    UUID REFERENCES users(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_login_at TIMESTAMPTZ
);
```

- [ ] Create `migrations/001_create_users.down.sql`:

```sql
DROP TABLE IF EXISTS users;
```

- [ ] Create `migrations/002_create_images.up.sql`:

```sql
CREATE TABLE images (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  image_id          TEXT NOT NULL UNIQUE,
  original_filename TEXT,
  thumbnail_key     TEXT,
  web_key           TEXT,
  original_key      TEXT,
  thumbnail_size    BIGINT,
  web_size          BIGINT,
  original_size     BIGINT,
  width             INTEGER,
  height            INTEGER,
  uploaded_by       UUID REFERENCES users(id),
  uploaded_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  published         BOOLEAN NOT NULL DEFAULT false,
  moderation_status TEXT NOT NULL DEFAULT 'pending'
    CHECK (moderation_status IN ('pending', 'approved', 'rejected')),
  date_type         TEXT CHECK (date_type IN ('exact', 'range', 'approximate')),
  exact_date        DATE,
  start_date        DATE,
  end_date          DATE,
  approx_year       INTEGER,
  approx_month      INTEGER,
  occasion_category TEXT CHECK (occasion_category IN (
    'birthday','wedding','graduation','holiday','vacation',
    'work_event','party','family_gathering','sports_event',
    'concert','conference','ceremony','casual','other'
  )),
  occasion_name     TEXT,
  exif              JSONB
);

CREATE INDEX ON images (uploaded_by);
CREATE INDEX ON images (uploaded_at DESC);
CREATE INDEX ON images (published);
CREATE INDEX ON images (date_type, exact_date);
```

- [ ] Create `migrations/002_create_images.down.sql`:

```sql
DROP TABLE IF EXISTS images;
```

- [ ] Create `migrations/003_create_tags.up.sql`:

```sql
CREATE TABLE tags (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL UNIQUE,
  usage_count INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by  UUID REFERENCES users(id)
);
```

- [ ] Create `migrations/003_create_tags.down.sql`:

```sql
DROP TABLE IF EXISTS tags;
```

- [ ] Create `migrations/004_create_image_tags.up.sql`:

```sql
CREATE TABLE image_tags (
  image_id UUID REFERENCES images(id) ON DELETE CASCADE,
  tag_id   UUID REFERENCES tags(id)   ON DELETE CASCADE,
  PRIMARY KEY (image_id, tag_id)
);

CREATE INDEX ON image_tags (tag_id);
```

- [ ] Create `migrations/004_create_image_tags.down.sql`:

```sql
DROP TABLE IF EXISTS image_tags;
```

- [ ] Create `migrations/005_create_image_people.up.sql`:

```sql
CREATE TABLE image_people (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
  name     TEXT NOT NULL
);

CREATE INDEX ON image_people (name);
```

- [ ] Create `migrations/005_create_image_people.down.sql`:

```sql
DROP TABLE IF EXISTS image_people;
```

- [ ] Create `migrations/006_create_magic_link_tokens.up.sql`:

```sql
CREATE TABLE magic_link_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ
);
```

- [ ] Create `migrations/006_create_magic_link_tokens.down.sql`:

```sql
DROP TABLE IF EXISTS magic_link_tokens;
```

- [ ] Verify by running migrations:

```bash
# Start postgres
docker run -d --name backend-postgres -p 5432:5432 \
  -e POSTGRES_USER=backend -e POSTGRES_PASSWORD=backend \
  -e POSTGRES_DB=imagedb postgres:16-alpine

# Install migrate CLI if needed
brew install golang-migrate

# Run migrations from services/backend/
migrate -path migrations -database "postgres://backend:backend@localhost:5432/imagedb?sslmode=disable" up

# Verify tables exist
docker exec backend-postgres psql -U backend -d imagedb -c "\dt"
```

Expected: 6 tables listed: `users`, `images`, `tags`, `image_tags`, `image_people`, `magic_link_tokens`.

**Commit:** `feat(backend): database migrations`

---

## Task 4: sqlc queries + generate

Write all 4 query files then run `sqlc generate`.

- [ ] Create `services/backend/internal/db/query/auth.sql`:

```sql
-- name: CreateUser :one
INSERT INTO users (email, name, role, status, invited_by)
VALUES (@email, @name, @role, @status, @invited_by)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = @email LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = @id LIMIT 1;

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = now() WHERE id = @id;

-- name: CreateMagicLinkToken :one
INSERT INTO magic_link_tokens (user_id, token_hash, expires_at)
VALUES (@user_id, @token_hash, @expires_at)
RETURNING *;

-- name: GetMagicLinkTokenByHash :one
SELECT * FROM magic_link_tokens WHERE token_hash = @token_hash LIMIT 1;

-- name: MarkTokenUsed :exec
UPDATE magic_link_tokens SET used_at = now() WHERE id = @id;
```

- [ ] Create `services/backend/internal/db/query/images.sql`:

```sql
-- name: CreateImage :one
INSERT INTO images (
  image_id, original_filename, thumbnail_key, web_key, original_key,
  thumbnail_size, web_size, original_size, width, height, uploaded_by, exif
) VALUES (
  @image_id, @original_filename, @thumbnail_key, @web_key, @original_key,
  @thumbnail_size, @web_size, @original_size, @width, @height, @uploaded_by, @exif
)
RETURNING *;

-- name: GetImageByID :one
SELECT * FROM images WHERE id = @id LIMIT 1;

-- name: GetImageByImageID :one
SELECT * FROM images WHERE image_id = @image_id LIMIT 1;

-- name: ListImages :many
SELECT * FROM images
WHERE (@occasion_category::text IS NULL OR occasion_category = @occasion_category)
ORDER BY uploaded_at DESC
LIMIT @lim OFFSET @off;

-- name: UpdateImage :one
UPDATE images SET
  published         = COALESCE(@published, published),
  date_type         = COALESCE(@date_type, date_type),
  exact_date        = COALESCE(@exact_date, exact_date),
  start_date        = COALESCE(@start_date, start_date),
  end_date          = COALESCE(@end_date, end_date),
  occasion_category = COALESCE(@occasion_category, occasion_category),
  occasion_name     = COALESCE(@occasion_name, occasion_name)
WHERE id = @id
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images WHERE id = @id;

-- name: CreateImagePerson :one
INSERT INTO image_people (image_id, name) VALUES (@image_id, @name) RETURNING *;

-- name: DeleteImagePeople :exec
DELETE FROM image_people WHERE image_id = @image_id;

-- name: ListImagePeople :many
SELECT * FROM image_people WHERE image_id = @image_id ORDER BY name;
```

- [ ] Create `services/backend/internal/db/query/tags.sql`:

```sql
-- name: CreateTag :one
INSERT INTO tags (name, created_by)
VALUES (@name, @created_by)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name;

-- name: SearchTags :many
SELECT * FROM tags
WHERE name ILIKE '%' || @query || '%'
ORDER BY usage_count DESC
LIMIT 10;

-- name: AddImageTag :exec
INSERT INTO image_tags (image_id, tag_id) VALUES (@image_id, @tag_id)
ON CONFLICT DO NOTHING;

-- name: RemoveImageTag :exec
DELETE FROM image_tags WHERE image_id = @image_id AND tag_id = @tag_id;

-- name: ListImageTags :many
SELECT t.* FROM tags t
JOIN image_tags it ON it.tag_id = t.id
WHERE it.image_id = @image_id
ORDER BY t.name;

-- name: IncrementTagUsage :exec
UPDATE tags SET usage_count = usage_count + 1 WHERE id = @id;

-- name: DecrementTagUsage :exec
UPDATE tags SET usage_count = GREATEST(0, usage_count - 1) WHERE id = @id;
```

- [ ] Create `services/backend/internal/db/query/users.sql`:

```sql
-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: UpdateUserRole :one
UPDATE users SET role = @role WHERE id = @id RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users SET status = @status WHERE id = @id RETURNING *;
```

- [ ] Install sqlc if needed and generate:

```bash
brew install sqlc
cd services/backend && sqlc generate
```

Expected: `internal/db/` populated with `db.go`, `models.go`, `querier.go`, `auth.sql.go`, `images.sql.go`, `tags.sql.go`, `users.sql.go`.

- [ ] Verify compilation:

```bash
cd services/backend && go build ./internal/db/...
```

Key generated types used in subsequent tasks:

```go
// models.go (generated)
type User struct {
    ID          pgtype.UUID
    Email       string
    Name        pgtype.Text
    Role        string
    Status      string
    InvitedBy   pgtype.UUID
    CreatedAt   pgtype.Timestamptz
    LastLoginAt pgtype.Timestamptz
}

type Image struct {
    ID               pgtype.UUID
    ImageID          string
    OriginalFilename pgtype.Text
    ThumbnailKey     pgtype.Text
    WebKey           pgtype.Text
    OriginalKey      pgtype.Text
    ThumbnailSize    pgtype.Int8
    WebSize          pgtype.Int8
    OriginalSize     pgtype.Int8
    Width            pgtype.Int4
    Height           pgtype.Int4
    UploadedBy       pgtype.UUID
    UploadedAt       pgtype.Timestamptz
    Published        bool
    ModerationStatus string
    DateType         pgtype.Text
    ExactDate        pgtype.Date
    StartDate        pgtype.Date
    EndDate          pgtype.Date
    ApproxYear       pgtype.Int4
    ApproxMonth      pgtype.Int4
    OccasionCategory pgtype.Text
    OccasionName     pgtype.Text
    Exif             []byte
}

type Tag struct {
    ID         pgtype.UUID
    Name       string
    UsageCount int32
    CreatedAt  pgtype.Timestamptz
    CreatedBy  pgtype.UUID
}

type MagicLinkToken struct {
    ID        pgtype.UUID
    UserID    pgtype.UUID
    TokenHash string
    ExpiresAt pgtype.Timestamptz
    UsedAt    pgtype.Timestamptz
}

type ImagePerson struct {
    ID      pgtype.UUID
    ImageID pgtype.UUID
    Name    string
}
```

**Commit:** `feat(backend): sqlc queries and generated db layer`

---

## Task 5: API errors + JWT middleware

- [ ] Create `services/backend/internal/api/errors.go`:

```go
package api

import (
    "encoding/json"
    "net/http"
)

type APIError struct {
    Type     string `json:"type"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail"`
    Instance string `json:"instance,omitempty"`
}

func (e APIError) Error() string { return e.Detail }

func writeError(w http.ResponseWriter, r *http.Request, e APIError) {
    e.Instance = r.URL.Path
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(e.Status)
    json.NewEncoder(w).Encode(e)
}

func errValidation(detail string) APIError {
    return APIError{Type: "validation_error", Title: "Validation Error", Status: http.StatusBadRequest, Detail: detail}
}

func errUnauthorized(detail string) APIError {
    return APIError{Type: "unauthorized", Title: "Unauthorized", Status: http.StatusUnauthorized, Detail: detail}
}

func errForbidden(detail string) APIError {
    return APIError{Type: "forbidden", Title: "Forbidden", Status: http.StatusForbidden, Detail: detail}
}

func errNotFound(detail string) APIError {
    return APIError{Type: "not_found", Title: "Not Found", Status: http.StatusNotFound, Detail: detail}
}

func errInternal(detail string) APIError {
    return APIError{Type: "internal_error", Title: "Internal Server Error", Status: http.StatusInternalServerError, Detail: detail}
}
```

- [ ] Create `services/backend/internal/api/middleware.go`:

```go
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
```

- [ ] Create `services/backend/internal/api/middleware_test.go`:

```go
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
```

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/api/... -run TestJWT -v
cd services/backend && go test ./internal/api/... -run TestRequireAdmin -v
```

Expected: all 6 tests pass.

**Commit:** `feat(backend): RFC 7807 errors and JWT middleware`

---

## Task 6: Mailer

- [ ] Create `services/backend/internal/mailer/mailer.go`:

```go
package mailer

import (
    "fmt"
    "log"
    "net/smtp"
)

// Mailer sends transactional emails.
type Mailer interface {
    SendMagicLink(toEmail, magicLinkURL string) error
}

// SMTPMailer sends emails via SMTP.
type SMTPMailer struct {
    host string
    port int
    from string
}

// NewSMTPMailer creates an SMTPMailer.
func NewSMTPMailer(host string, port int, from string) *SMTPMailer {
    return &SMTPMailer{host: host, port: port, from: from}
}

// SendMagicLink sends the magic link email via SMTP.
func (m *SMTPMailer) SendMagicLink(to, url string) error {
    addr := fmt.Sprintf("%s:%d", m.host, m.port)
    msg := []byte(fmt.Sprintf(
        "To: %s\r\nFrom: %s\r\nSubject: Your sign-in link\r\n\r\nClick to sign in: %s\r\n",
        to, m.from, url,
    ))
    return smtp.SendMail(addr, nil, m.from, []string{to}, msg)
}

// LogMailer logs magic links instead of sending emails. Use in DEV_MODE.
type LogMailer struct{}

// NewLogMailer creates a LogMailer.
func NewLogMailer() *LogMailer { return &LogMailer{} }

// SendMagicLink logs the magic link URL to stdout.
func (m *LogMailer) SendMagicLink(to, url string) error {
    log.Printf("[DEV] Magic link for %s: %s", to, url)
    return nil
}
```

- [ ] Create `services/backend/internal/mailer/mailer_test.go`:

```go
package mailer_test

import (
    "testing"

    "github.com/leahgarrett/image-management-system/services/backend/internal/mailer"
)

func TestLogMailer_DoesNotError(t *testing.T) {
    m := mailer.NewLogMailer()
    err := m.SendMagicLink("test@example.com", "http://localhost:3000/auth/verify?token=abc123")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
}
```

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/mailer/...
```

Expected:
```
ok      github.com/leahgarrett/image-management-system/services/backend/internal/mailer
```

**Commit:** `feat(backend): mailer interface with SMTP and dev log implementations`

---

## Task 7: Auth handlers

- [ ] Create `services/backend/internal/api/handlers_auth.go`:

```go
package api

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/leahgarrett/image-management-system/services/backend/internal/db"
    "github.com/leahgarrett/image-management-system/services/backend/internal/mailer"
)

// Handlers holds all handler dependencies.
type Handlers struct {
    q         db.Querier
    mailer    mailer.Mailer
    appURL    string
    jwtSecret string
}

// NewHandlers constructs a Handlers instance.
func NewHandlers(q db.Querier, m mailer.Mailer, appURL, jwtSecret string) *Handlers {
    return &Handlers{q: q, mailer: m, appURL: appURL, jwtSecret: jwtSecret}
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
        if errors.Is(err, pgx.ErrNoRows) {
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

    expiresAt := pgtype.Timestamptz{Time: time.Now().Add(15 * time.Minute), Valid: true}
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
        if errors.Is(err, pgx.ErrNoRows) {
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
    if time.Now().After(record.ExpiresAt.Time) {
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
```

- [ ] Create `services/backend/internal/api/handlers_auth_test.go`:

```go
package api_test

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/jackc/pgx/v5/pgtype"
    "github.com/leahgarrett/image-management-system/services/backend/internal/api"
    "github.com/leahgarrett/image-management-system/services/backend/internal/db"
    "github.com/leahgarrett/image-management-system/services/backend/internal/mailer"
)

// mockQuerier implements db.Querier for tests. All methods not under test
// return zero values or nil errors by default.
type mockQuerier struct {
    // Auth fields
    getUserByEmailFn        func(ctx context.Context, email string) (db.User, error)
    getUserByIDFn           func(ctx context.Context, id pgtype.UUID) (db.User, error)
    listUsersFn             func(ctx context.Context) ([]db.User, error)
    createUserFn            func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
    updateUserLastLoginFn   func(ctx context.Context, id pgtype.UUID) error
    createMagicLinkTokenFn  func(ctx context.Context, arg db.CreateMagicLinkTokenParams) (db.MagicLinkToken, error)
    getMagicLinkTokenByHashFn func(ctx context.Context, tokenHash string) (db.MagicLinkToken, error)
    markTokenUsedFn         func(ctx context.Context, id pgtype.UUID) error

    // Image fields
    createImageFn      func(ctx context.Context, arg db.CreateImageParams) (db.Image, error)
    getImageByIDFn     func(ctx context.Context, id pgtype.UUID) (db.Image, error)
    getImageByImageIDFn func(ctx context.Context, imageID string) (db.Image, error)
    listImagesFn       func(ctx context.Context, arg db.ListImagesParams) ([]db.Image, error)
    updateImageFn      func(ctx context.Context, arg db.UpdateImageParams) (db.Image, error)
    deleteImageFn      func(ctx context.Context, id pgtype.UUID) error
    createImagePersonFn func(ctx context.Context, arg db.CreateImagePersonParams) (db.ImagePerson, error)
    deleteImagePeopleFn func(ctx context.Context, imageID pgtype.UUID) error
    listImagePeopleFn  func(ctx context.Context, imageID pgtype.UUID) ([]db.ImagePerson, error)

    // Tag fields
    createTagFn        func(ctx context.Context, arg db.CreateTagParams) (db.Tag, error)
    listTagsFn         func(ctx context.Context) ([]db.Tag, error)
    searchTagsFn       func(ctx context.Context, query string) ([]db.Tag, error)
    addImageTagFn      func(ctx context.Context, arg db.AddImageTagParams) error
    removeImageTagFn   func(ctx context.Context, arg db.RemoveImageTagParams) error
    listImageTagsFn    func(ctx context.Context, imageID pgtype.UUID) ([]db.Tag, error)
    incrementTagUsageFn func(ctx context.Context, id pgtype.UUID) error
    decrementTagUsageFn func(ctx context.Context, id pgtype.UUID) error

    // User admin fields
    updateUserRoleFn   func(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error)
    updateUserStatusFn func(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error)
}

func (m *mockQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
    if m.getUserByEmailFn != nil {
        return m.getUserByEmailFn(ctx, email)
    }
    return db.User{}, nil
}
func (m *mockQuerier) GetUserByID(ctx context.Context, id pgtype.UUID) (db.User, error) {
    if m.getUserByIDFn != nil {
        return m.getUserByIDFn(ctx, id)
    }
    return db.User{}, nil
}
func (m *mockQuerier) ListUsers(ctx context.Context) ([]db.User, error) {
    if m.listUsersFn != nil {
        return m.listUsersFn(ctx)
    }
    return []db.User{}, nil
}
func (m *mockQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
    if m.createUserFn != nil {
        return m.createUserFn(ctx, arg)
    }
    return db.User{Email: arg.Email, Role: arg.Role, Status: arg.Status}, nil
}
func (m *mockQuerier) UpdateUserLastLogin(ctx context.Context, id pgtype.UUID) error {
    if m.updateUserLastLoginFn != nil {
        return m.updateUserLastLoginFn(ctx, id)
    }
    return nil
}
func (m *mockQuerier) CreateMagicLinkToken(ctx context.Context, arg db.CreateMagicLinkTokenParams) (db.MagicLinkToken, error) {
    if m.createMagicLinkTokenFn != nil {
        return m.createMagicLinkTokenFn(ctx, arg)
    }
    return db.MagicLinkToken{TokenHash: arg.TokenHash}, nil
}
func (m *mockQuerier) GetMagicLinkTokenByHash(ctx context.Context, tokenHash string) (db.MagicLinkToken, error) {
    if m.getMagicLinkTokenByHashFn != nil {
        return m.getMagicLinkTokenByHashFn(ctx, tokenHash)
    }
    return db.MagicLinkToken{}, nil
}
func (m *mockQuerier) MarkTokenUsed(ctx context.Context, id pgtype.UUID) error {
    if m.markTokenUsedFn != nil {
        return m.markTokenUsedFn(ctx, id)
    }
    return nil
}
func (m *mockQuerier) CreateImage(ctx context.Context, arg db.CreateImageParams) (db.Image, error) {
    if m.createImageFn != nil {
        return m.createImageFn(ctx, arg)
    }
    return db.Image{ImageID: arg.ImageID}, nil
}
func (m *mockQuerier) GetImageByID(ctx context.Context, id pgtype.UUID) (db.Image, error) {
    if m.getImageByIDFn != nil {
        return m.getImageByIDFn(ctx, id)
    }
    return db.Image{}, nil
}
func (m *mockQuerier) GetImageByImageID(ctx context.Context, imageID string) (db.Image, error) {
    if m.getImageByImageIDFn != nil {
        return m.getImageByImageIDFn(ctx, imageID)
    }
    return db.Image{}, nil
}
func (m *mockQuerier) ListImages(ctx context.Context, arg db.ListImagesParams) ([]db.Image, error) {
    if m.listImagesFn != nil {
        return m.listImagesFn(ctx, arg)
    }
    return []db.Image{}, nil
}
func (m *mockQuerier) UpdateImage(ctx context.Context, arg db.UpdateImageParams) (db.Image, error) {
    if m.updateImageFn != nil {
        return m.updateImageFn(ctx, arg)
    }
    return db.Image{}, nil
}
func (m *mockQuerier) DeleteImage(ctx context.Context, id pgtype.UUID) error {
    if m.deleteImageFn != nil {
        return m.deleteImageFn(ctx, id)
    }
    return nil
}
func (m *mockQuerier) CreateImagePerson(ctx context.Context, arg db.CreateImagePersonParams) (db.ImagePerson, error) {
    if m.createImagePersonFn != nil {
        return m.createImagePersonFn(ctx, arg)
    }
    return db.ImagePerson{Name: arg.Name}, nil
}
func (m *mockQuerier) DeleteImagePeople(ctx context.Context, imageID pgtype.UUID) error {
    if m.deleteImagePeopleFn != nil {
        return m.deleteImagePeopleFn(ctx, imageID)
    }
    return nil
}
func (m *mockQuerier) ListImagePeople(ctx context.Context, imageID pgtype.UUID) ([]db.ImagePerson, error) {
    if m.listImagePeopleFn != nil {
        return m.listImagePeopleFn(ctx, imageID)
    }
    return []db.ImagePerson{}, nil
}
func (m *mockQuerier) CreateTag(ctx context.Context, arg db.CreateTagParams) (db.Tag, error) {
    if m.createTagFn != nil {
        return m.createTagFn(ctx, arg)
    }
    return db.Tag{Name: arg.Name}, nil
}
func (m *mockQuerier) ListTags(ctx context.Context) ([]db.Tag, error) {
    if m.listTagsFn != nil {
        return m.listTagsFn(ctx)
    }
    return []db.Tag{}, nil
}
func (m *mockQuerier) SearchTags(ctx context.Context, query string) ([]db.Tag, error) {
    if m.searchTagsFn != nil {
        return m.searchTagsFn(ctx, query)
    }
    return []db.Tag{}, nil
}
func (m *mockQuerier) AddImageTag(ctx context.Context, arg db.AddImageTagParams) error {
    if m.addImageTagFn != nil {
        return m.addImageTagFn(ctx, arg)
    }
    return nil
}
func (m *mockQuerier) RemoveImageTag(ctx context.Context, arg db.RemoveImageTagParams) error {
    if m.removeImageTagFn != nil {
        return m.removeImageTagFn(ctx, arg)
    }
    return nil
}
func (m *mockQuerier) ListImageTags(ctx context.Context, imageID pgtype.UUID) ([]db.Tag, error) {
    if m.listImageTagsFn != nil {
        return m.listImageTagsFn(ctx, imageID)
    }
    return []db.Tag{}, nil
}
func (m *mockQuerier) IncrementTagUsage(ctx context.Context, id pgtype.UUID) error {
    if m.incrementTagUsageFn != nil {
        return m.incrementTagUsageFn(ctx, id)
    }
    return nil
}
func (m *mockQuerier) DecrementTagUsage(ctx context.Context, id pgtype.UUID) error {
    if m.decrementTagUsageFn != nil {
        return m.decrementTagUsageFn(ctx, id)
    }
    return nil
}
func (m *mockQuerier) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
    if m.updateUserRoleFn != nil {
        return m.updateUserRoleFn(ctx, arg)
    }
    return db.User{}, nil
}
func (m *mockQuerier) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error) {
    if m.updateUserStatusFn != nil {
        return m.updateUserStatusFn(ctx, arg)
    }
    return db.User{}, nil
}

// --- Auth handler tests ---

func newTestHandlers(q db.Querier) *api.Handlers {
    return api.NewHandlers(q, mailer.NewLogMailer(), "http://localhost:3000", testSecret)
}

func TestLogin_CreatesTokenAndSendsLink(t *testing.T) {
    q := &mockQuerier{
        getUserByEmailFn: func(ctx context.Context, email string) (db.User, error) {
            return db.User{}, pgx.ErrNoRows
        },
    }
    h := newTestHandlers(q)
    body := `{"email":"test@example.com"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    h.Login(rr, req)
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
    }
    var resp map[string]string
    json.NewDecoder(rr.Body).Decode(&resp)
    if resp["message"] != "Magic link sent" {
        t.Errorf("unexpected message: %s", resp["message"])
    }
}

func TestVerify_ValidToken_SetsCookie(t *testing.T) {
    validUUID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
    q := &mockQuerier{
        getMagicLinkTokenByHashFn: func(ctx context.Context, tokenHash string) (db.MagicLinkToken, error) {
            return db.MagicLinkToken{
                ID:        validUUID,
                UserID:    validUUID,
                TokenHash: tokenHash,
                ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Minute), Valid: true},
            }, nil
        },
        getUserByIDFn: func(ctx context.Context, id pgtype.UUID) (db.User, error) {
            return db.User{ID: validUUID, Email: "test@example.com", Role: "contributor"}, nil
        },
    }
    h := newTestHandlers(q)
    req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify?token=abc123deadbeef", nil)
    rr := httptest.NewRecorder()
    h.Verify(rr, req)
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
    }
    cookies := rr.Result().Cookies()
    found := false
    for _, c := range cookies {
        if c.Name == "auth_token" && c.HttpOnly {
            found = true
        }
    }
    if !found {
        t.Error("expected httpOnly auth_token cookie to be set")
    }
}

func TestVerify_ExpiredToken_Returns401(t *testing.T) {
    validUUID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
    q := &mockQuerier{
        getMagicLinkTokenByHashFn: func(ctx context.Context, tokenHash string) (db.MagicLinkToken, error) {
            return db.MagicLinkToken{
                ID:        validUUID,
                UserID:    validUUID,
                TokenHash: tokenHash,
                ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
            }, nil
        },
    }
    h := newTestHandlers(q)
    req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify?token=abc123", nil)
    rr := httptest.NewRecorder()
    h.Verify(rr, req)
    if rr.Code != http.StatusUnauthorized {
        t.Errorf("expected 401, got %d", rr.Code)
    }
}

func TestVerify_UsedToken_Returns401(t *testing.T) {
    validUUID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
    usedAt := pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}
    q := &mockQuerier{
        getMagicLinkTokenByHashFn: func(ctx context.Context, tokenHash string) (db.MagicLinkToken, error) {
            return db.MagicLinkToken{
                ID:        validUUID,
                UserID:    validUUID,
                TokenHash: tokenHash,
                ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Minute), Valid: true},
                UsedAt:    usedAt,
            }, nil
        },
    }
    h := newTestHandlers(q)
    req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify?token=abc123", nil)
    rr := httptest.NewRecorder()
    h.Verify(rr, req)
    if rr.Code != http.StatusUnauthorized {
        t.Errorf("expected 401, got %d", rr.Code)
    }
}

func TestLogout_ClearsCookie(t *testing.T) {
    h := newTestHandlers(&mockQuerier{})
    req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
    rr := httptest.NewRecorder()
    h.Logout(rr, req)
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rr.Code)
    }
    cookies := rr.Result().Cookies()
    for _, c := range cookies {
        if c.Name == "auth_token" && c.MaxAge != 0 {
            t.Errorf("expected MaxAge=0 for cleared cookie, got %d", c.MaxAge)
        }
    }
}
```

Note: `handlers_auth_test.go` imports `"github.com/jackc/pgx/v5"` for `pgx.ErrNoRows`.

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/api/... -run TestLogin -v
cd services/backend && go test ./internal/api/... -run TestVerify -v
cd services/backend && go test ./internal/api/... -run TestLogout -v
```

Expected: all 5 auth tests pass.

**Commit:** `feat(backend): magic link auth handlers`

---

## Task 8: Image registration handler

- [ ] Create `services/backend/internal/api/handlers_images.go` (initial — POST only):

```go
package api

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/jackc/pgx/v5/pgtype"
    "github.com/leahgarrett/image-management-system/services/backend/internal/db"
)

type imageMetadata struct {
    CaptureDate string `json:"captureDate"`
    CameraMake  string `json:"cameraMake"`
    CameraModel string `json:"cameraModel"`
}

type registerImageRequest struct {
    ImageID          string        `json:"imageId"`
    OriginalFilename string        `json:"originalFilename"`
    ThumbnailKey     string        `json:"thumbnailKey"`
    WebKey           string        `json:"webKey"`
    OriginalKey      string        `json:"originalKey"`
    ThumbnailSize    int64         `json:"thumbnailSize"`
    WebSize          int64         `json:"webSize"`
    OriginalSize     int64         `json:"originalSize"`
    Width            int32         `json:"width"`
    Height           int32         `json:"height"`
    Metadata         imageMetadata `json:"metadata"`
}

type imageResponse struct {
    ID               string        `json:"id"`
    ImageID          string        `json:"imageId"`
    OriginalFilename string        `json:"originalFilename"`
    ThumbnailKey     string        `json:"thumbnailKey"`
    WebKey           string        `json:"webKey"`
    OriginalKey      string        `json:"originalKey"`
    ThumbnailSize    int64         `json:"thumbnailSize"`
    WebSize          int64         `json:"webSize"`
    OriginalSize     int64         `json:"originalSize"`
    Width            int32         `json:"width"`
    Height           int32         `json:"height"`
    UploadedAt       time.Time     `json:"uploadedAt"`
    Published        bool          `json:"published"`
    People           []string      `json:"people"`
    Tags             []string      `json:"tags"`
    DateType         string        `json:"dateType,omitempty"`
    ExactDate        string        `json:"exactDate,omitempty"`
    OccasionCategory string        `json:"occasionCategory,omitempty"`
    OccasionName     string        `json:"occasionName,omitempty"`
    Metadata         imageMetadata `json:"metadata"`
}

func imageToResponse(img db.Image, people []db.ImagePerson, tags []db.Tag) imageResponse {
    resp := imageResponse{
        ID:               img.ID.String(),
        ImageID:          img.ImageID,
        OriginalFilename: img.OriginalFilename.String,
        ThumbnailKey:     img.ThumbnailKey.String,
        WebKey:           img.WebKey.String,
        OriginalKey:      img.OriginalKey.String,
        ThumbnailSize:    img.ThumbnailSize.Int64,
        WebSize:          img.WebSize.Int64,
        OriginalSize:     img.OriginalSize.Int64,
        Width:            img.Width.Int32,
        Height:           img.Height.Int32,
        UploadedAt:       img.UploadedAt.Time,
        Published:        img.Published,
        DateType:         img.DateType.String,
        OccasionCategory: img.OccasionCategory.String,
        OccasionName:     img.OccasionName.String,
        People:           []string{},
        Tags:             []string{},
    }
    if img.ExactDate.Valid {
        resp.ExactDate = img.ExactDate.Time.Format("2006-01-02")
    }
    for _, p := range people {
        resp.People = append(resp.People, p.Name)
    }
    for _, tg := range tags {
        resp.Tags = append(resp.Tags, tg.Name)
    }
    return resp
}

// RegisterImage handles POST /api/v1/images.
func (h *Handlers) RegisterImage(w http.ResponseWriter, r *http.Request) {
    var req registerImageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, r, errValidation("invalid request body"))
        return
    }
    if req.ImageID == "" {
        writeError(w, r, errValidation("imageId is required"))
        return
    }

    userID, ok := UserIDFromContext(r.Context())
    if !ok {
        writeError(w, r, errUnauthorized("missing user context"))
        return
    }

    metaJSON, err := json.Marshal(req.Metadata)
    if err != nil {
        writeError(w, r, errInternal("failed to marshal metadata"))
        return
    }

    // Parse uploaded_by UUID.
    var uploadedBy pgtype.UUID
    if err := uploadedBy.Scan(userID); err != nil {
        writeError(w, r, errInternal("invalid user ID"))
        return
    }

    params := db.CreateImageParams{
        ImageID:          req.ImageID,
        OriginalFilename: pgtype.Text{String: req.OriginalFilename, Valid: req.OriginalFilename != ""},
        ThumbnailKey:     pgtype.Text{String: req.ThumbnailKey, Valid: req.ThumbnailKey != ""},
        WebKey:           pgtype.Text{String: req.WebKey, Valid: req.WebKey != ""},
        OriginalKey:      pgtype.Text{String: req.OriginalKey, Valid: req.OriginalKey != ""},
        ThumbnailSize:    pgtype.Int8{Int64: req.ThumbnailSize, Valid: req.ThumbnailSize > 0},
        WebSize:          pgtype.Int8{Int64: req.WebSize, Valid: req.WebSize > 0},
        OriginalSize:     pgtype.Int8{Int64: req.OriginalSize, Valid: req.OriginalSize > 0},
        Width:            pgtype.Int4{Int32: req.Width, Valid: req.Width > 0},
        Height:           pgtype.Int4{Int32: req.Height, Valid: req.Height > 0},
        UploadedBy:       uploadedBy,
        Exif:             metaJSON,
    }

    img, err := h.q.CreateImage(r.Context(), params)
    if err != nil {
        writeError(w, r, errInternal("failed to create image"))
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(imageToResponse(img, nil, nil))
}
```

- [ ] Create `services/backend/internal/api/handlers_images_test.go` (initial):

```go
package api_test

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/leahgarrett/image-management-system/services/backend/internal/api"
    "github.com/leahgarrett/image-management-system/services/backend/internal/db"
)

func TestRegisterImage_ValidRequest_Returns201(t *testing.T) {
    q := &mockQuerier{
        createImageFn: func(ctx context.Context, arg db.CreateImageParams) (db.Image, error) {
            return db.Image{ImageID: arg.ImageID}, nil
        },
    }
    h := newTestHandlers(q)
    body := `{
        "imageId": "img-abc-123",
        "originalFilename": "photo.jpg",
        "thumbnailKey": "user/img/thumbnail.jpg",
        "webKey": "user/img/web.jpg",
        "originalKey": "user/img/original.jpg",
        "thumbnailSize": 30000,
        "webSize": 300000,
        "originalSize": 15000000,
        "width": 4032,
        "height": 3024,
        "metadata": {"captureDate": "2024-06-15T10:30:00Z", "cameraMake": "Apple", "cameraModel": "iPhone 14 Pro"}
    }`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/images", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.RegisterImage(rr, req)
    if rr.Code != http.StatusCreated {
        t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
    }
}

func TestRegisterImage_MissingImageID_Returns400(t *testing.T) {
    h := newTestHandlers(&mockQuerier{})
    body := `{"originalFilename": "photo.jpg"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/images", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.RegisterImage(rr, req)
    if rr.Code != http.StatusBadRequest {
        t.Errorf("expected 400, got %d", rr.Code)
    }
}

func TestRegisterImage_SetsUploadedByFromContext(t *testing.T) {
    const testUserID = "00000000-0000-0000-0000-000000000001"
    var capturedUploadedBy string
    q := &mockQuerier{
        createImageFn: func(ctx context.Context, arg db.CreateImageParams) (db.Image, error) {
            capturedUploadedBy = arg.UploadedBy.String()
            return db.Image{ImageID: arg.ImageID}, nil
        },
    }
    h := newTestHandlers(q)
    body := `{"imageId": "img-test-001", "metadata": {}}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/images", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    ctx := api.ContextWithUserID(req.Context(), testUserID)
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.RegisterImage(rr, req)
    if rr.Code != http.StatusCreated {
        t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
    }
    if capturedUploadedBy != testUserID {
        t.Errorf("expected uploadedBy %s, got %s", testUserID, capturedUploadedBy)
    }
}
```

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/api/... -run TestRegisterImage -v
```

Expected: all 3 tests pass.

**Commit:** `feat(backend): image registration handler`

---

## Task 9: Image list + get handlers

- [ ] Append to `services/backend/internal/api/handlers_images.go`:

```go
type paginationResponse struct {
    Total   int  `json:"total"`
    Limit   int  `json:"limit"`
    Offset  int  `json:"offset"`
    HasMore bool `json:"hasMore"`
}

type listImagesResponse struct {
    Data       []imageResponse    `json:"data"`
    Pagination paginationResponse `json:"pagination"`
}

// ListImages handles GET /api/v1/images.
func (h *Handlers) ListImages(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    occasion := q.Get("occasion")
    limit := 20
    offset := 0
    if v := q.Get("limit"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            if n > 100 {
                n = 100
            }
            limit = n
        }
    }
    if v := q.Get("offset"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n >= 0 {
            offset = n
        }
    }

    var occasionParam pgtype.Text
    if occasion != "" {
        occasionParam = pgtype.Text{String: occasion, Valid: true}
    }

    images, err := h.q.ListImages(r.Context(), db.ListImagesParams{
        OccasionCategory: occasionParam,
        Lim:              int32(limit),
        Off:              int32(offset),
    })
    if err != nil {
        writeError(w, r, errInternal("failed to list images"))
        return
    }

    data := make([]imageResponse, 0, len(images))
    for _, img := range images {
        people, _ := h.q.ListImagePeople(r.Context(), img.ID)
        tags, _ := h.q.ListImageTags(r.Context(), img.ID)
        data = append(data, imageToResponse(img, people, tags))
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(listImagesResponse{
        Data: data,
        Pagination: paginationResponse{
            Total:   len(data),
            Limit:   limit,
            Offset:  offset,
            HasMore: len(data) == limit,
        },
    })
}

// GetImage handles GET /api/v1/images/{id}.
func (h *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]

    var id pgtype.UUID
    if err := id.Scan(idStr); err != nil {
        writeError(w, r, errValidation("invalid image id"))
        return
    }

    img, err := h.q.GetImageByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            writeError(w, r, errNotFound("image not found"))
            return
        }
        writeError(w, r, errInternal("failed to get image"))
        return
    }

    people, _ := h.q.ListImagePeople(r.Context(), img.ID)
    tags, _ := h.q.ListImageTags(r.Context(), img.ID)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(imageToResponse(img, people, tags))
}
```

Add required imports to `handlers_images.go`:

```go
import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/leahgarrett/image-management-system/services/backend/internal/db"
)
```

- [ ] Append to `services/backend/internal/api/handlers_images_test.go`:

```go
func TestListImages_Returns200WithPagination(t *testing.T) {
    q := &mockQuerier{
        listImagesFn: func(ctx context.Context, arg db.ListImagesParams) ([]db.Image, error) {
            return []db.Image{{ImageID: "img-001"}, {ImageID: "img-002"}}, nil
        },
    }
    h := newTestHandlers(q)
    req := httptest.NewRequest(http.MethodGet, "/api/v1/images?limit=10&offset=0", nil)
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.ListImages(rr, req)
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
    }
    var resp struct {
        Data       []interface{}       `json:"data"`
        Pagination map[string]interface{} `json:"pagination"`
    }
    json.NewDecoder(rr.Body).Decode(&resp)
    if len(resp.Data) != 2 {
        t.Errorf("expected 2 images, got %d", len(resp.Data))
    }
    if resp.Pagination == nil {
        t.Error("expected pagination in response")
    }
}

func TestGetImage_Returns200WithPeopleAndTags(t *testing.T) {
    imageUUID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
    q := &mockQuerier{
        getImageByIDFn: func(ctx context.Context, id pgtype.UUID) (db.Image, error) {
            return db.Image{ID: imageUUID, ImageID: "img-001"}, nil
        },
        listImagePeopleFn: func(ctx context.Context, imageID pgtype.UUID) ([]db.ImagePerson, error) {
            return []db.ImagePerson{{Name: "Alice"}, {Name: "Bob"}}, nil
        },
        listImageTagsFn: func(ctx context.Context, imageID pgtype.UUID) ([]db.Tag, error) {
            return []db.Tag{{Name: "vacation"}}, nil
        },
    }
    h := newTestHandlers(q)
    req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+imageUUID.String(), nil)
    req = mux.SetURLVars(req, map[string]string{"id": imageUUID.String()})
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.GetImage(rr, req)
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
    }
    var resp map[string]interface{}
    json.NewDecoder(rr.Body).Decode(&resp)
    people, _ := resp["people"].([]interface{})
    if len(people) != 2 {
        t.Errorf("expected 2 people, got %v", people)
    }
    tags, _ := resp["tags"].([]interface{})
    if len(tags) != 1 {
        t.Errorf("expected 1 tag, got %v", tags)
    }
}

func TestGetImage_NotFound_Returns404(t *testing.T) {
    q := &mockQuerier{
        getImageByIDFn: func(ctx context.Context, id pgtype.UUID) (db.Image, error) {
            return db.Image{}, pgx.ErrNoRows
        },
    }
    h := newTestHandlers(q)
    id := "00000000-0000-0000-0000-000000000099"
    req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+id, nil)
    req = mux.SetURLVars(req, map[string]string{"id": id})
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.GetImage(rr, req)
    if rr.Code != http.StatusNotFound {
        t.Errorf("expected 404, got %d", rr.Code)
    }
}
```

Note: tests use `mux.SetURLVars` from `github.com/gorilla/mux` to inject path variables in unit tests without a full router.

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/api/... -run TestListImages -v
cd services/backend && go test ./internal/api/... -run TestGetImage -v
```

Expected: all 3 tests pass.

**Commit:** `feat(backend): image list and get handlers`

---

## Task 10: Image update + delete handlers

- [ ] Append to `services/backend/internal/api/handlers_images.go`:

```go
type updateImageRequest struct {
    People           []string `json:"people"`
    Tags             []string `json:"tags"`
    DateType         string   `json:"dateType"`
    ExactDate        string   `json:"exactDate"`
    OccasionCategory string   `json:"occasionCategory"`
    OccasionName     string   `json:"occasionName"`
    Published        *bool    `json:"published"`
}

// UpdateImage handles PATCH /api/v1/images/{id}.
func (h *Handlers) UpdateImage(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]

    var id pgtype.UUID
    if err := id.Scan(idStr); err != nil {
        writeError(w, r, errValidation("invalid image id"))
        return
    }

    // Verify image exists.
    img, err := h.q.GetImageByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            writeError(w, r, errNotFound("image not found"))
            return
        }
        writeError(w, r, errInternal("failed to get image"))
        return
    }

    var req updateImageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, r, errValidation("invalid request body"))
        return
    }

    ctx := r.Context()

    // Update people: delete all, then re-insert.
    if req.People != nil {
        if err := h.q.DeleteImagePeople(ctx, img.ID); err != nil {
            writeError(w, r, errInternal("failed to update people"))
            return
        }
        for _, name := range req.People {
            if _, err := h.q.CreateImagePerson(ctx, db.CreateImagePersonParams{
                ImageID: img.ID,
                Name:    name,
            }); err != nil {
                writeError(w, r, errInternal("failed to add person"))
                return
            }
        }
    }

    // Update tags: diff existing vs requested.
    if req.Tags != nil {
        existingTags, err := h.q.ListImageTags(ctx, img.ID)
        if err != nil {
            writeError(w, r, errInternal("failed to list tags"))
            return
        }

        existingSet := map[string]db.Tag{}
        for _, t := range existingTags {
            existingSet[t.Name] = t
        }
        requestedSet := map[string]struct{}{}
        for _, name := range req.Tags {
            requestedSet[name] = struct{}{}
        }

        // Add new tags.
        for name := range requestedSet {
            if _, found := existingSet[name]; !found {
                tag, err := h.q.CreateTag(ctx, db.CreateTagParams{Name: name})
                if err != nil {
                    writeError(w, r, errInternal("failed to create tag"))
                    return
                }
                if err := h.q.AddImageTag(ctx, db.AddImageTagParams{ImageID: img.ID, TagID: tag.ID}); err != nil {
                    writeError(w, r, errInternal("failed to add image tag"))
                    return
                }
                if err := h.q.IncrementTagUsage(ctx, tag.ID); err != nil {
                    writeError(w, r, errInternal("failed to increment tag usage"))
                    return
                }
            }
        }

        // Remove tags no longer in request.
        for name, tag := range existingSet {
            if _, found := requestedSet[name]; !found {
                if err := h.q.RemoveImageTag(ctx, db.RemoveImageTagParams{ImageID: img.ID, TagID: tag.ID}); err != nil {
                    writeError(w, r, errInternal("failed to remove image tag"))
                    return
                }
                if err := h.q.DecrementTagUsage(ctx, tag.ID); err != nil {
                    writeError(w, r, errInternal("failed to decrement tag usage"))
                    return
                }
            }
        }
    }

    // Build UpdateImage params.
    params := db.UpdateImageParams{ID: img.ID}
    if req.Published != nil {
        params.Published = pgtype.Bool{Bool: *req.Published, Valid: true}
    }
    if req.DateType != "" {
        params.DateType = pgtype.Text{String: req.DateType, Valid: true}
    }
    if req.ExactDate != "" {
        t, err := time.Parse("2006-01-02", req.ExactDate)
        if err == nil {
            params.ExactDate = pgtype.Date{Time: t, Valid: true}
        }
    }
    if req.OccasionCategory != "" {
        params.OccasionCategory = pgtype.Text{String: req.OccasionCategory, Valid: true}
    }
    if req.OccasionName != "" {
        params.OccasionName = pgtype.Text{String: req.OccasionName, Valid: true}
    }

    updatedImg, err := h.q.UpdateImage(ctx, params)
    if err != nil {
        writeError(w, r, errInternal("failed to update image"))
        return
    }

    people, _ := h.q.ListImagePeople(ctx, updatedImg.ID)
    tags, _ := h.q.ListImageTags(ctx, updatedImg.ID)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(imageToResponse(updatedImg, people, tags))
}

// DeleteImage handles DELETE /api/v1/images/{id}.
func (h *Handlers) DeleteImage(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]

    var id pgtype.UUID
    if err := id.Scan(idStr); err != nil {
        writeError(w, r, errValidation("invalid image id"))
        return
    }

    if err := h.q.DeleteImage(r.Context(), id); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            writeError(w, r, errNotFound("image not found"))
            return
        }
        writeError(w, r, errInternal("failed to delete image"))
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

- [ ] Append to `services/backend/internal/api/handlers_images_test.go`:

```go
func TestUpdateImage_UpdatesPeopleAndTags(t *testing.T) {
    imageUUID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
    q := &mockQuerier{
        getImageByIDFn: func(ctx context.Context, id pgtype.UUID) (db.Image, error) {
            return db.Image{ID: imageUUID, ImageID: "img-001"}, nil
        },
        updateImageFn: func(ctx context.Context, arg db.UpdateImageParams) (db.Image, error) {
            return db.Image{ID: imageUUID, ImageID: "img-001"}, nil
        },
        listImageTagsFn: func(ctx context.Context, imageID pgtype.UUID) ([]db.Tag, error) {
            return []db.Tag{}, nil // no existing tags
        },
        createTagFn: func(ctx context.Context, arg db.CreateTagParams) (db.Tag, error) {
            return db.Tag{Name: arg.Name}, nil
        },
    }
    h := newTestHandlers(q)
    body := `{"people": ["Alice", "Bob"], "tags": ["vacation", "beach"]}`
    req := httptest.NewRequest(http.MethodPatch, "/api/v1/images/"+imageUUID.String(), bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    req = mux.SetURLVars(req, map[string]string{"id": imageUUID.String()})
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.UpdateImage(rr, req)
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
    }
}

func TestUpdateImage_NotFound_Returns404(t *testing.T) {
    q := &mockQuerier{
        getImageByIDFn: func(ctx context.Context, id pgtype.UUID) (db.Image, error) {
            return db.Image{}, pgx.ErrNoRows
        },
    }
    h := newTestHandlers(q)
    id := "00000000-0000-0000-0000-000000000099"
    req := httptest.NewRequest(http.MethodPatch, "/api/v1/images/"+id, bytes.NewBufferString(`{}`))
    req = mux.SetURLVars(req, map[string]string{"id": id})
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.UpdateImage(rr, req)
    if rr.Code != http.StatusNotFound {
        t.Errorf("expected 404, got %d", rr.Code)
    }
}

func TestDeleteImage_Returns204(t *testing.T) {
    imageUUID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
    q := &mockQuerier{
        deleteImageFn: func(ctx context.Context, id pgtype.UUID) error {
            return nil
        },
    }
    h := newTestHandlers(q)
    req := httptest.NewRequest(http.MethodDelete, "/api/v1/images/"+imageUUID.String(), nil)
    req = mux.SetURLVars(req, map[string]string{"id": imageUUID.String()})
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.DeleteImage(rr, req)
    if rr.Code != http.StatusNoContent {
        t.Errorf("expected 204, got %d", rr.Code)
    }
}
```

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/api/... -run TestUpdateImage -v
cd services/backend && go test ./internal/api/... -run TestDeleteImage -v
```

Expected: all 3 tests pass.

**Commit:** `feat(backend): image update and delete handlers`

---

## Task 11: Tags handlers

- [ ] Create `services/backend/internal/api/handlers_tags.go`:

```go
package api

import (
    "encoding/json"
    "net/http"
)

type tagResponse struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    UsageCount int32  `json:"usageCount"`
}

// ListTags handles GET /api/v1/tags.
func (h *Handlers) ListTags(w http.ResponseWriter, r *http.Request) {
    tags, err := h.q.ListTags(r.Context())
    if err != nil {
        writeError(w, r, errInternal("failed to list tags"))
        return
    }

    resp := make([]tagResponse, 0, len(tags))
    for _, t := range tags {
        resp = append(resp, tagResponse{
            ID:         t.ID.String(),
            Name:       t.Name,
            UsageCount: t.UsageCount,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{"tags": resp})
}

// TagSuggestions handles GET /api/v1/tags/suggestions?q=xxx.
func (h *Handlers) TagSuggestions(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    if query == "" {
        writeError(w, r, errValidation("q query parameter is required"))
        return
    }

    tags, err := h.q.SearchTags(r.Context(), query)
    if err != nil {
        writeError(w, r, errInternal("failed to search tags"))
        return
    }

    resp := make([]tagResponse, 0, len(tags))
    for _, t := range tags {
        resp = append(resp, tagResponse{
            ID:         t.ID.String(),
            Name:       t.Name,
            UsageCount: t.UsageCount,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{"tags": resp})
}
```

- [ ] Create `services/backend/internal/api/handlers_tags_test.go`:

```go
package api_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/leahgarrett/image-management-system/services/backend/internal/api"
    "github.com/leahgarrett/image-management-system/services/backend/internal/db"
)

func TestListTags_Returns200(t *testing.T) {
    q := &mockQuerier{
        listTagsFn: func(ctx context.Context) ([]db.Tag, error) {
            return []db.Tag{
                {Name: "vacation", UsageCount: 5},
                {Name: "beach", UsageCount: 3},
            }, nil
        },
    }
    h := newTestHandlers(q)
    req := httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.ListTags(rr, req)
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
    }
    var resp struct {
        Tags []tagResponse `json:"tags"`
    }
    json.NewDecoder(rr.Body).Decode(&resp)
    if len(resp.Tags) != 2 {
        t.Errorf("expected 2 tags, got %d", len(resp.Tags))
    }
}

type tagResponse struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    UsageCount int32  `json:"usageCount"`
}

func TestSuggestions_FiltersCorrectly(t *testing.T) {
    q := &mockQuerier{
        searchTagsFn: func(ctx context.Context, query string) ([]db.Tag, error) {
            if query != "vac" {
                return []db.Tag{}, nil
            }
            return []db.Tag{{Name: "vacation", UsageCount: 5}}, nil
        },
    }
    h := newTestHandlers(q)
    req := httptest.NewRequest(http.MethodGet, "/api/v1/tags/suggestions?q=vac", nil)
    ctx := api.ContextWithUserID(req.Context(), "00000000-0000-0000-0000-000000000001")
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    h.TagSuggestions(rr, req)
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
    }
    var resp struct {
        Tags []tagResponse `json:"tags"`
    }
    json.NewDecoder(rr.Body).Decode(&resp)
    if len(resp.Tags) != 1 || resp.Tags[0].Name != "vacation" {
        t.Errorf("unexpected suggestions: %v", resp.Tags)
    }
}
```

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/api/... -run TestListTags -v
cd services/backend && go test ./internal/api/... -run TestSuggestions -v
```

Expected: both tests pass.

**Commit:** `feat(backend): tags handlers`

---

## Task 12: Users handlers

- [ ] Create `services/backend/internal/api/handlers_users.go`:

```go
package api

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/jackc/pgx/v5/pgtype"
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
        CreatedAt: u.CreatedAt.Time.Format(time.RFC3339),
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

    inviterID, _ := UserIDFromContext(r.Context())
    var invitedBy pgtype.UUID
    invitedBy.Scan(inviterID)

    user, err := h.q.CreateUser(r.Context(), db.CreateUserParams{
        Email:     req.Email,
        Name:      pgtype.Text{String: req.Name, Valid: req.Name != ""},
        Role:      req.Role,
        Status:    "invited",
        InvitedBy: invitedBy,
    })
    if err != nil {
        writeError(w, r, errInternal("failed to create user"))
        return
    }

    // Generate and send magic link.
    rawBytes := make([]byte, 32)
    rand.Read(rawBytes)
    rawToken := hex.EncodeToString(rawBytes)

    import_sha256 := func() string {
        // inline to avoid extra import block confusion — use crypto/sha256
        h256 := sha256.Sum256([]byte(rawToken))
        return hex.EncodeToString(h256[:])
    }
    // NOTE: actual file should import crypto/sha256 at top of file.
    tokenHash := import_sha256()

    expiresAt := pgtype.Timestamptz{Time: time.Now().Add(72 * time.Hour), Valid: true}
    h.q.CreateMagicLinkToken(r.Context(), db.CreateMagicLinkTokenParams{
        UserID:    user.ID,
        TokenHash: tokenHash,
        ExpiresAt: expiresAt,
    })

    magicLinkURL := fmt.Sprintf("%s/auth/verify?token=%s", h.appURL, rawToken)
    h.mailer.SendMagicLink(req.Email, magicLinkURL)

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

    var id pgtype.UUID
    if err := id.Scan(idStr); err != nil {
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
```

Note: the `import_sha256` inline function in `InviteUser` is illustrative. In the actual file, add `"crypto/sha256"` to the import block at the top and call it directly:

```go
// At top of file:
import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    ...
)

// In InviteUser:
hash := sha256.Sum256([]byte(rawToken))
tokenHash := hex.EncodeToString(hash[:])
```

- [ ] Create `services/backend/internal/api/handlers_users_test.go`:

```go
package api_test

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gorilla/mux"
    "github.com/jackc/pgx/v5/pgtype"
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
    // RequireAdmin middleware blocks the request before ListUsers is called.
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
    userUUID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
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
```

- [ ] Run tests:

```bash
cd services/backend && go test ./internal/api/... -run TestListUsers -v
cd services/backend && go test ./internal/api/... -run TestInviteUser -v
cd services/backend && go test ./internal/api/... -run TestUpdateRole -v
```

Expected: all 4 tests pass.

**Commit:** `feat(backend): users handlers`

---

## Task 13: HTTP server

- [ ] Create `services/backend/internal/api/server.go`:

```go
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
```

- [ ] Verify the package compiles:

```bash
cd services/backend && go build ./internal/api/...
```

Expected: no errors.

**Commit:** `feat(backend): gorilla/mux router`

---

## Task 14: Wire up main.go

- [ ] Create `services/backend/main.go`:

```go
package main

import (
    "context"
    "embed"
    "errors"
    "log"
    "net/http"
    "time"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
    "github.com/golang-migrate/migrate/v4/source/iofs"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/leahgarrett/image-management-system/services/backend/internal/api"
    "github.com/leahgarrett/image-management-system/services/backend/internal/config"
    "github.com/leahgarrett/image-management-system/services/backend/internal/db"
    "github.com/leahgarrett/image-management-system/services/backend/internal/mailer"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("config: %v", err)
    }

    // Run database migrations.
    src, err := iofs.New(migrationsFS, "migrations")
    if err != nil {
        log.Fatalf("migrations source: %v", err)
    }
    m, err := migrate.NewWithSourceInstance("iofs", src, cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("migrate init: %v", err)
    }
    if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
        log.Fatalf("migrate up: %v", err)
    }
    log.Println("migrations applied")

    // Connect to PostgreSQL.
    ctx := context.Background()
    pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("db connect: %v", err)
    }
    defer pool.Close()

    if err := pool.Ping(ctx); err != nil {
        log.Fatalf("db ping: %v", err)
    }

    // Build dependencies.
    queries := db.New(pool)

    var mlr mailer.Mailer
    if cfg.DevMode {
        mlr = mailer.NewLogMailer()
        log.Println("using log mailer (DEV_MODE=true)")
    } else {
        mlr = mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
    }

    handlers := api.NewHandlers(queries, mlr, cfg.AppURL, cfg.JWTSecret)
    router := api.NewRouter(handlers, cfg.JWTSecret)

    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      router,
        WriteTimeout: 30 * time.Second,
        ReadTimeout:  15 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    log.Printf("backend service starting on :%s", cfg.Port)
    log.Fatal(srv.ListenAndServe())
}
```

- [ ] Build and run:

```bash
cd services/backend && go build -o backend-service . && ./backend-service
```

- [ ] Smoke test (in a separate terminal with DEV_MODE=true and a running postgres):

```bash
curl http://localhost:8081/health
```

Expected response:
```json
{"status":"ok"}
```

**Commit:** `feat(backend): wire all dependencies in main.go`

---

## Task 15: Dockerfile

- [ ] Create `services/backend/Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o backend-service .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/backend-service .
EXPOSE 8081
CMD ["./backend-service"]
```

Note: no CGO is needed for this service (unlike the ingestion service which may use cgo for image processing). The pure Go pgx driver and pure Go JWT library require no native libraries.

- [ ] Build and smoke test the Docker image:

```bash
cd services/backend

# Build
docker build -t backend-service:local .

# Run with env file
docker run --rm \
  -e DATABASE_URL="postgres://backend:backend@host.docker.internal:5432/imagedb?sslmode=disable" \
  -e JWT_SECRET="changeme" \
  -e APP_URL="http://localhost:3000" \
  -e DEV_MODE="true" \
  -p 8081:8081 \
  backend-service:local

# In a second terminal:
curl http://localhost:8081/health
```

Expected:
```json
{"status":"ok"}
```

- [ ] Full test suite:

```bash
cd services/backend && go test ./...
```

Expected: all packages pass with no failures.

**Commit:** `feat(backend): Dockerfile`

---

## Summary

| Task | File(s) | Tests |
|------|---------|-------|
| 1 | go.mod, sqlc.yaml, .env.example | go mod download |
| 2 | internal/config/config.go | 5 config tests |
| 3 | migrations/001–006 *.sql | docker + migrate verify |
| 4 | internal/db/query/*.sql → generated | sqlc generate + go build |
| 5 | internal/api/errors.go, middleware.go | 6 middleware tests |
| 6 | internal/mailer/mailer.go | 1 mailer test |
| 7 | internal/api/handlers_auth.go | 5 auth handler tests |
| 8 | internal/api/handlers_images.go (POST) | 3 register tests |
| 9 | handlers_images.go (GET list/single) | 3 list/get tests |
| 10 | handlers_images.go (PATCH/DELETE) | 3 update/delete tests |
| 11 | internal/api/handlers_tags.go | 2 tags tests |
| 12 | internal/api/handlers_users.go | 4 users tests |
| 13 | internal/api/server.go | go build verify |
| 14 | main.go | curl /health |
| 15 | Dockerfile | docker build + curl |
