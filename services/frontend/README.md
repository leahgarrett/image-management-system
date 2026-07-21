# Frontend

React SPA for browsing and searching the image archive. Handles authentication, image browsing with filters, and (eventually) uploads.

## Stack

- **React 18** + **TypeScript** with Vite
- **Mantine v7** component library
- **Zustand** for auth and filter state
- **React Router v6** with nested protected routes
- **Vitest** + Testing Library for unit tests

## Getting started

### Prerequisites

- Node.js 18+
- Backend service running on port 8081
- Ingestion service running on port 8080 (for uploads)

### Run locally

```bash
npm install
npm run dev
```

Vite dev server starts on [http://localhost:3000](http://localhost:3000) and proxies API requests:
- `/api/v1/ingest/*` → port 8080 (ingestion service)
- `/api/*` → port 8081 (backend service)

### Run tests

```bash
npm test
```

## Authentication

Uses magic link (passwordless) auth via the backend service. On first load, the app calls `GET /api/v1/users/me` to check for an existing session. If the cookie is valid, the user lands on the browse page; otherwise they're redirected to `/login`.

Auth state is held in Zustand (`authStore`). The `isInitialising` flag prevents a flash-to-login on hard reload while the session check is in flight.

## Project structure

```
src/
├── api/
│   └── client.ts          # apiFetch wrapper — credentials: include, RFC 7807 error handling
├── features/
│   ├── auth/
│   │   ├── authStore.ts   # Zustand auth state (user, isInitialising, requestMagicLink, logout)
│   │   ├── LoginPage.tsx  # Email input → magic link request
│   │   └── VerifyPage.tsx # Reads ?token=, calls /auth/verify, redirects
│   └── images/
│       ├── types.ts        # ImageSummary, ImageListResponse
│       ├── filterStore.ts  # Zustand filter state (tags, people, occasion, pagination)
│       ├── FilterBar.tsx   # Tags / People / Occasion inputs + Clear filters
│       ├── BrowsePage.tsx  # Fetches image list, renders FilterBar + ImageGrid + pagination
│       ├── ImageGrid.tsx   # Responsive grid (discriminated union props: loading | error | images)
│       └── ImageCard.tsx   # Thumbnail placeholder, filename, tag badges, people text
└── shared/
    ├── AppShell.tsx        # Mantine AppShell with header and Sign out
    └── ProtectedRoute.tsx  # Redirects unauthenticated users; shows loader while initialising
```

## Known limitations

- Tags and people filter params are accepted by the UI but not yet applied in the backend SQL query — results are unfiltered until that is implemented.
- Pagination `total` from the backend currently reflects page size, not the true total count.
- No upload UI — images are ingested via the ingestion service directly.
