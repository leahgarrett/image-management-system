# Backend Service

REST API for the image management system. Handles image metadata, user authentication, and tag management.

## Stack

- **Go** with gorilla/mux
- **PostgreSQL** with sqlc (type-safe queries)
- **golang-migrate** for schema migrations
- **JWT** + magic link passwordless authentication

## Getting started

### Prerequisites

- Go 1.21+
- PostgreSQL

### Run locally

```bash
cp .env.example .env
# Edit .env with your database URL and secrets

go run .
```

With `DEV_MODE=true`, magic links are logged to stdout instead of sent via email — no SMTP setup needed.

### Run tests

```bash
go test ./...
```

## Configuration

All configuration is via environment variables.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_SECRET` | Yes | — | Secret for signing JWTs |
| `APP_URL` | Yes | — | Frontend base URL (used in magic link emails) |
| `PORT` | No | `8081` | Port to listen on |
| `DEV_MODE` | No | `false` | Logs magic links to stdout; disables SMTP requirement |
| `SMTP_HOST` | When `DEV_MODE=false` | — | SMTP server hostname |
| `SMTP_PORT` | No | `587` | SMTP server port |
| `SMTP_FROM` | When `DEV_MODE=false` | — | From address for outbound emails |

## API

Base path: `/api/v1`

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/auth/login` | — | Request a magic link |
| `GET` | `/auth/verify?token=` | — | Verify token, issue JWT cookie |
| `POST` | `/auth/logout` | — | Clear auth cookie |

### Images

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/images` | JWT | List images (filterable) |
| `POST` | `/images` | JWT | Register a new image |
| `GET` | `/images/:id` | JWT | Get a single image |
| `PATCH` | `/images/:id` | JWT | Update image metadata |
| `DELETE` | `/images/:id` | JWT | Delete an image |

**List filters** (query params): `tags`, `people`, `dateFrom`, `dateTo`, `occasionCategory`, `limit`, `offset`

### Tags

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/tags` | JWT | List all tags |
| `GET` | `/tags/suggestions?q=` | JWT | Search tags by prefix |

### Users (admin only)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/users` | JWT + admin | List all users |
| `POST` | `/users/invite` | JWT + admin | Invite a new user |
| `PATCH` | `/users/:id/role` | JWT + admin | Update a user's role |

### Health

```
GET /health
```

### Error format

All errors follow [RFC 7807](https://datatracker.ietf.org/doc/html/rfc7807):

```json
{
  "type": "validation_error",
  "title": "Validation Error",
  "status": 400,
  "detail": "email is required",
  "instance": "/api/v1/auth/login"
}
```

## Docker

```bash
docker build -t backend .
docker run -p 8081:8081 --env-file .env backend
```

## Project structure

```
.
├── main.go                  # Entry point: wires config, DB, mailer, server
├── migrations/              # SQL migration files (embedded at build time)
├── internal/
│   ├── api/                 # HTTP handlers, middleware, router
│   ├── config/              # Environment-based config loading
│   ├── db/                  # sqlc-generated query code
│   └── mailer/              # Mailer interface (SMTP + dev log implementations)
└── sqlc.yaml                # sqlc configuration
```
