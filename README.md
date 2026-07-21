# Image Management System
A simple, community-built media library for browsing and search.

## About

This project is in early development. We're building an image archive system for families and small groups to preserve important memories with their stories and context.

## Services

| Service | Language | Port | Description |
|---------|----------|------|-------------|
| [frontend](services/frontend/) | React / TypeScript | 3000 | SPA for browsing and searching images |
| [backend](services/backend/) | Go | 8081 | REST API — auth, image metadata, tags |
| [ingestion](services/ingestion/) | Go | 8080 | Async image upload and processing |

## Running locally

Each service has its own README with setup instructions. The typical local setup:

1. Start PostgreSQL and run backend migrations
2. `cd services/backend && go run .`
3. `cd services/ingestion && go run .`
4. `cd services/frontend && npm install && npm run dev`

The frontend dev server proxies API requests to the backend and ingestion services automatically.

## Project Structure

- **docs/** - Project documentation including meeting notes, product research, and planning materials
- **services/** - Individual services (frontend, backend, ingestion)
