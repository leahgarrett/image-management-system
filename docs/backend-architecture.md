# Backend Architecture

_Note:_ This document outlines the technical decisions for the backend API: language, database, authentication approach, and API design patterns.

## Decision History

| Date | Decision | Reason |
|------|----------|--------|
| Initial | Node.js + Express + MongoDB + Mongoose | Familiar stack, flexible document model for image metadata |
| 2026-04-22 | **Changed to Go + PostgreSQL + sqlc** | Consistency with ingestion service (already Go); relational model is a better fit for structured metadata with filtering; sqlc provides better type safety than Mongoose |

---

## Backend Language & Framework

### Decision: Go with gorilla/mux

**Rationale:**
- Consistent with the ingestion service — one language, one toolchain, shared patterns
- gorilla/mux already proven in the ingestion service
- Strong typing catches errors at compile time
- Good performance for a CRUD + search workload

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| **Go + gorilla/mux** | Consistent with ingestion service, type-safe, mature | More boilerplate than Node.js frameworks |
| **Node.js + Express** | Original choice; large ecosystem, fast to write | Different language from ingestion service; weaker typing |
| **Go + chi** | Lightweight, idiomatic Go | Less familiar than gorilla/mux given existing usage |

**Documentation:** https://github.com/gorilla/mux

---

## Database

### Decision: PostgreSQL

**Why PostgreSQL over MongoDB:**

The image metadata model is more relational than it first appears:
- Users, images, comments, and tags have clear relationships with referential integrity requirements
- Filtering queries (by person, date range, occasion, tag) are what SQL was built for
- ACID transactions keep tag usage counts and image status consistent
- JSONB columns handle the genuinely flexible parts (EXIF data, dateRange) without giving up the whole schema

MongoDB's flexible document model was over-engineering for a schema that is mostly well-defined.

### Database Access: sqlc

**Rationale:**
- Generates type-safe Go code directly from SQL queries — no ORM abstraction layer
- Queries are plain SQL, easy to read and optimise
- Compile-time safety: if a query doesn't match the schema, the build fails
- No magic — what you write is what runs

### Migrations: golang-migrate

**Rationale:**
- Simple file-based migrations (up/down SQL files)
- CLI tool for running migrations in CI and locally
- Works with PostgreSQL natively

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| **PostgreSQL + sqlc** | Type-safe, plain SQL, great Go support | More upfront schema design, migration management required |
| **MongoDB + Mongoose** | Original choice; flexible schema, rich middleware hooks | Weaker query typing, less suited to relational filtering |
| **PostgreSQL + GORM** | ORM familiarity, less boilerplate | Magic behaviour, harder to optimise queries |
| **PostgreSQL + pgx raw** | Maximum control | Too much boilerplate for CRUD operations |

**Documentation:**
- pgx: https://github.com/jackc/pgx
- sqlc: https://sqlc.dev/
- golang-migrate: https://github.com/golang-migrate/migrate

---

## Database Schema

### Tables

#### images
```sql
CREATE TABLE images (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  image_id      TEXT NOT NULL UNIQUE,
  original_filename TEXT,
  thumbnail_key TEXT,
  web_key       TEXT,
  original_key  TEXT,
  thumbnail_size BIGINT,
  web_size       BIGINT,
  original_size  BIGINT,
  width         INTEGER,
  height        INTEGER,
  uploaded_by   UUID REFERENCES users(id),
  uploaded_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  published     BOOLEAN NOT NULL DEFAULT false,
  moderation_status TEXT NOT NULL DEFAULT 'pending'
    CHECK (moderation_status IN ('pending', 'approved', 'rejected')),
  date_type     TEXT CHECK (date_type IN ('exact', 'range', 'approximate')),
  exact_date    DATE,
  start_date    DATE,
  end_date      DATE,
  approx_year   INTEGER,
  approx_month  INTEGER,
  occasion_category TEXT CHECK (occasion_category IN (
    'birthday','wedding','graduation','holiday','vacation',
    'work_event','party','family_gathering','sports_event',
    'concert','conference','ceremony','casual','other'
  )),
  occasion_name TEXT,
  exif          JSONB
);
```

#### users
```sql
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

#### tags
```sql
CREATE TABLE tags (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL UNIQUE,
  usage_count INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by  UUID REFERENCES users(id)
);
```

#### image_tags
```sql
CREATE TABLE image_tags (
  image_id UUID REFERENCES images(id) ON DELETE CASCADE,
  tag_id   UUID REFERENCES tags(id)   ON DELETE CASCADE,
  PRIMARY KEY (image_id, tag_id)
);
```

#### image_people
```sql
CREATE TABLE image_people (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
  name     TEXT NOT NULL
);
```

#### comments
```sql
CREATE TABLE comments (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  image_id          UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
  user_id           UUID NOT NULL REFERENCES users(id),
  text              TEXT NOT NULL,
  parent_comment_id UUID REFERENCES comments(id),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  moderation_status TEXT NOT NULL DEFAULT 'approved'
    CHECK (moderation_status IN ('approved', 'flagged', 'hidden'))
);
```

#### magic_link_tokens
```sql
CREATE TABLE magic_link_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ
);
```

### Key Indexes
```sql
CREATE INDEX ON images (uploaded_by);
CREATE INDEX ON images (uploaded_at DESC);
CREATE INDEX ON images (published);
CREATE INDEX ON images (date_type, exact_date);
CREATE INDEX ON image_people (name);
CREATE INDEX ON image_tags (tag_id);
CREATE INDEX ON comments (image_id);
```

### Design Principles
- dateRange stored as flat columns (not nested JSON) for efficient range queries
- EXIF stored as JSONB — flexible, rarely queried directly
- Tags normalised into a separate table to support usage counts and suggestions
- People embedded per-image (no global people registry in V1)

---

## Authentication & Authorization

### Decision: JWT with Magic Link (unchanged)

Magic link auth remains the right choice — passwordless is better UX for family/non-technical users.

**Magic Link Flow:**
1. User enters email
2. Backend generates a short-lived token (15 min), stores `hash(token)` in `magic_link_tokens`
3. Email sent with magic link: `https://app.example.com/auth/verify?token=xyz`
4. User clicks link → backend validates token hash, marks `used_at`, issues JWT
5. JWT stored in httpOnly cookie

**JWT Payload:**
```json
{
  "userId": "507f1f77bcf86cd799439011",
  "email": "user@example.com",
  "role": "contributor",
  "permissions": ["images.view", "images.upload", "images.tag"],
  "iat": 1642546800,
  "exp": 1642633200
}
```

**Go Libraries:**
- `golang-jwt/jwt/v5` — JWT generation/verification (already used in ingestion service)
- `net/smtp` or AWS SES SDK — email delivery
- `golang.org/x/crypto` — token hashing (bcrypt or SHA-256)
- `golang.org/x/time/rate` — rate limiting magic link requests

---

## API Design

### REST Structure

**Base URL:** `/api/v1`

```
GET    /api/v1/images              # List images (with filters)
GET    /api/v1/images/:id          # Get single image
POST   /api/v1/images              # Register image (after ingestion completes)
PATCH  /api/v1/images/:id          # Update metadata/tags
DELETE /api/v1/images/:id          # Delete image

GET    /api/v1/images/:id/comments # Get comments for image
POST   /api/v1/images/:id/comments # Add comment

GET    /api/v1/tags                # Get all tags
GET    /api/v1/tags/suggestions    # Get suggested tags

POST   /api/v1/auth/login          # Request magic link
GET    /api/v1/auth/verify         # Verify magic link token
POST   /api/v1/auth/logout         # Invalidate token

GET    /api/v1/users               # List users (admin only)
POST   /api/v1/users/invite        # Invite new user (admin only)
PATCH  /api/v1/users/:id/role      # Update user role (admin only)
```

### Error Handling

**Standardized Error Response (RFC 7807) — consistent with ingestion service:**
```json
{
  "type": "validation_error",
  "title": "Validation Error",
  "status": 400,
  "detail": "Image file size exceeds maximum allowed (15MB)",
  "instance": "/api/v1/images"
}
```

### Query Parameters for Filtering

```
GET /api/v1/images?tags=vacation,beach&people=John&dateFrom=2024-01-01&dateTo=2024-12-31&limit=20&offset=0
```

**Response Format:**
```json
{
  "data": [...],
  "pagination": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "hasMore": true
  }
}
```

---

## Summary

| Concern | Decision |
|---------|----------|
| Language | Go |
| HTTP framework | gorilla/mux |
| Database | PostgreSQL |
| Query generation | sqlc |
| Migrations | golang-migrate |
| Auth | JWT + magic link |
| JWT library | golang-jwt/jwt/v5 |
| Error format | RFC 7807 (consistent with ingestion service) |
