# Frontend Core — Auth + Browse Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a working React frontend that lets authenticated users browse and filter images, using the existing Go backend API.

**Architecture:** Vite + React + TypeScript SPA. Mantine v7 for all UI components (TagsInput, Dropzone, DatePickerInput, AppShell). Zustand for auth and filter state. Magic link authentication only — no password, no OAuth. JWT is stored in an httpOnly cookie set by the backend; the frontend never sees the token.

**Tech Stack:** React 18, TypeScript, Vite, Mantine v7, Zustand, React Router v6

---

## File Structure

```
services/frontend/
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
└── src/
    ├── main.tsx                       # MantineProvider, Notifications, BrowserRouter
    ├── App.tsx                        # Routes: /, /login, /auth/verify
    ├── theme.ts                       # Mantine theme overrides (minimal)
    ├── api/
    │   └── client.ts                  # fetch wrapper: credentials:include, RFC7807 errors
    ├── features/
    │   ├── auth/
    │   │   ├── authStore.ts           # Zustand: user | null, requestMagicLink, fetchCurrentUser, logout
    │   │   ├── LoginPage.tsx          # Email form → POST /api/v1/auth/login
    │   │   └── VerifyPage.tsx         # ?token= → GET /api/v1/auth/verify → redirect /
    │   └── images/
    │       ├── filterStore.ts         # Zustand: tags[], people, occasion, page
    │       ├── BrowsePage.tsx         # FilterBar + ImageGrid, fetches /api/v1/images
    │       ├── FilterBar.tsx          # Mantine inputs wired to filterStore
    │       ├── ImageGrid.tsx          # Mantine SimpleGrid of ImageCards
    │       └── ImageCard.tsx          # Single image card: thumbnail, filename, tags
    └── shared/
        ├── AppShell.tsx               # Mantine AppShell: nav header, user display, logout
        └── ProtectedRoute.tsx         # Redirects to /login if !user
```

**Backend prerequisite (Task 0):** `GET /api/v1/users/me` is needed so `fetchCurrentUser` can identify the logged-in user after the page loads. This endpoint is not currently in the backend — it must be added before Task 3 (auth store) can be completed.

---

## Task 0: Add GET /api/v1/users/me to the backend

**Files:**
- Modify: `services/backend/internal/api/handlers_users.go`
- Modify: `services/backend/internal/api/server.go`

- [ ] **Step 1: Add the handler to handlers_users.go**

Add this function at the end of `services/backend/internal/api/handlers_users.go`:

```go
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, r, errUnauthorized("missing auth context"))
		return
	}

	userID, _ := claims["userId"].(string)
	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":    userID,
		"email": email,
		"role":  role,
	})
}
```

- [ ] **Step 2: Register the route in server.go**

In `services/backend/internal/api/server.go`, inside the authenticated `api` subrouter block (after the tags routes), add:

```go
api.HandleFunc("/users/me", h.Me).Methods(http.MethodGet)
```

The authenticated block should now include:

```go
api.HandleFunc("/images", h.ListImages).Methods(http.MethodGet)
api.HandleFunc("/images", h.RegisterImage).Methods(http.MethodPost)
api.HandleFunc("/images/{id}", h.GetImage).Methods(http.MethodGet)
api.HandleFunc("/images/{id}", h.UpdateImage).Methods(http.MethodPatch)
api.HandleFunc("/images/{id}", h.DeleteImage).Methods(http.MethodDelete)

api.HandleFunc("/tags", h.ListTags).Methods(http.MethodGet)
api.HandleFunc("/tags/suggestions", h.TagSuggestions).Methods(http.MethodGet)

api.HandleFunc("/users/me", h.Me).Methods(http.MethodGet)
```

- [ ] **Step 3: Verify it compiles**

```bash
cd services/backend && go build ./...
```

Expected: no output (clean build)

- [ ] **Step 4: Commit**

```bash
git add services/backend/internal/api/handlers_users.go services/backend/internal/api/server.go
git commit -m "feat: add GET /api/v1/users/me endpoint"
```

---

## Task 1: Vite + React scaffold

**Files:**
- Create: `services/frontend/package.json`
- Create: `services/frontend/tsconfig.json`
- Create: `services/frontend/tsconfig.node.json`
- Create: `services/frontend/vite.config.ts`
- Create: `services/frontend/index.html`
- Create: `services/frontend/src/main.tsx`
- Create: `services/frontend/src/App.tsx`
- Create: `services/frontend/src/theme.ts`

- [ ] **Step 1: Create the frontend directory and package.json**

```bash
mkdir -p services/frontend/src
```

Create `services/frontend/package.json`:

```json
{
  "name": "image-management-frontend",
  "version": "0.0.1",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "@mantine/core": "^7.17.0",
    "@mantine/dates": "^7.17.0",
    "@mantine/dropzone": "^7.17.0",
    "@mantine/hooks": "^7.17.0",
    "@mantine/notifications": "^7.17.0",
    "dayjs": "^1.11.13",
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-router-dom": "^6.30.0",
    "zustand": "^5.0.4"
  },
  "devDependencies": {
    "@types/react": "^18.3.22",
    "@types/react-dom": "^18.3.7",
    "@vitejs/plugin-react": "^4.5.0",
    "typescript": "^5.8.3",
    "vite": "^6.3.4"
  }
}
```

- [ ] **Step 2: Create TypeScript config files**

Create `services/frontend/tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": { "@/*": ["./src/*"] }
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

Create `services/frontend/tsconfig.node.json`:

```json
{
  "compilerOptions": {
    "composite": true,
    "skipLibCheck": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true
  },
  "include": ["vite.config.ts"]
}
```

- [ ] **Step 3: Create vite.config.ts**

Create `services/frontend/vite.config.ts`:

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  server: {
    port: 3000,
    proxy: {
      '/api/v1/ingest': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
  build: { sourcemap: true },
})
```

Note: `/api/v1/ingest` must come before `/api` — Vite matches proxies in order, most-specific first.

- [ ] **Step 4: Create index.html**

Create `services/frontend/index.html`:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Family Photos</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 5: Create theme.ts**

Create `services/frontend/src/theme.ts`:

```typescript
import { createTheme } from '@mantine/core'

export const theme = createTheme({})
```

- [ ] **Step 6: Create main.tsx**

Create `services/frontend/src/main.tsx`:

```tsx
import '@mantine/core/styles.css'
import '@mantine/dates/styles.css'
import '@mantine/dropzone/styles.css'
import '@mantine/notifications/styles.css'
import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { MantineProvider } from '@mantine/core'
import { Notifications } from '@mantine/notifications'
import { App } from './App'
import { theme } from './theme'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <MantineProvider theme={theme}>
        <Notifications />
        <App />
      </MantineProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
```

- [ ] **Step 7: Create App.tsx (placeholder routes)**

Create `services/frontend/src/App.tsx`:

```tsx
import { Routes, Route, Navigate } from 'react-router-dom'

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<div>Login placeholder</div>} />
      <Route path="/auth/verify" element={<div>Verify placeholder</div>} />
      <Route path="/" element={<div>Browse placeholder</div>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
```

- [ ] **Step 8: Install dependencies and verify dev server starts**

```bash
cd services/frontend && npm install
npm run build
```

Expected: TypeScript compiles, Vite bundles without error. The `dist/` directory is created.

- [ ] **Step 9: Commit**

```bash
git add services/frontend/
git commit -m "feat: scaffold Vite + React + Mantine frontend"
```

---

## Task 2: API client

**Files:**
- Create: `services/frontend/src/api/client.ts`

- [ ] **Step 1: Write a failing test**

This module is pure TypeScript logic — test with a mock fetch. Create `services/frontend/src/api/client.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { apiFetch, ApiError } from './client'

describe('apiFetch', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  it('returns parsed JSON on 2xx', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValue(
      new Response(JSON.stringify({ id: '1' }), { status: 200 })
    )
    const result = await apiFetch<{ id: string }>('/api/v1/images')
    expect(result.id).toBe('1')
  })

  it('throws ApiError with status and detail on non-2xx', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValue(
      new Response(
        JSON.stringify({ type: 'not_found', detail: 'Image not found', status: 404 }),
        { status: 404, headers: { 'Content-Type': 'application/json' } }
      )
    )
    await expect(apiFetch('/api/v1/images/x')).rejects.toBeInstanceOf(ApiError)
  })

  it('always sends credentials: include', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValue(new Response('{}', { status: 200 }))
    await apiFetch('/api/v1/images')
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/images',
      expect.objectContaining({ credentials: 'include' })
    )
  })
})
```

To run tests, add vitest to devDependencies first:

```bash
cd services/frontend && npm install -D vitest @vitest/ui jsdom @testing-library/react @testing-library/jest-dom
```

Add to `package.json` scripts:

```json
"test": "vitest run",
"test:watch": "vitest"
```

Add to `vite.config.ts` (inside `defineConfig`):

```typescript
test: {
  environment: 'jsdom',
  globals: true,
},
```

Run: `npm test`

Expected: FAIL — `apiFetch` and `ApiError` not defined

- [ ] **Step 2: Create src/api/client.ts**

```typescript
export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly detail: string,
    public readonly type: string = 'error',
  ) {
    super(detail)
    this.name = 'ApiError'
  }
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const res = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...init.headers,
    },
  })

  if (!res.ok) {
    let detail = res.statusText
    let type = 'error'
    try {
      const body = await res.json()
      detail = body.detail ?? body.message ?? detail
      type = body.type ?? type
    } catch {
      // non-JSON error body — use statusText
    }
    throw new ApiError(res.status, detail, type)
  }

  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}
```

- [ ] **Step 3: Run tests**

```bash
cd services/frontend && npm test
```

Expected: 3 tests pass

- [ ] **Step 4: Commit**

```bash
git add services/frontend/src/api/ services/frontend/src/api/client.test.ts services/frontend/package.json services/frontend/vite.config.ts
git commit -m "feat: add API client with credentials and RFC7807 error handling"
```

---

## Task 3: Auth store

**Files:**
- Create: `services/frontend/src/features/auth/authStore.ts`

- [ ] **Step 1: Write failing test**

Create `services/frontend/src/features/auth/authStore.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuthStore } from './authStore'

vi.mock('@/api/client', () => ({
  apiFetch: vi.fn(),
  ApiError: class ApiError extends Error {
    constructor(public status: number, public detail: string) { super(detail) }
  },
}))

import { apiFetch } from '@/api/client'

beforeEach(() => {
  useAuthStore.setState({ user: null })
  vi.clearAllMocks()
})

describe('fetchCurrentUser', () => {
  it('sets user when /users/me succeeds', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ id: 'u1', email: 'a@b.com', role: 'admin' })
    await useAuthStore.getState().fetchCurrentUser()
    expect(useAuthStore.getState().user).toEqual({ id: 'u1', email: 'a@b.com', role: 'admin' })
  })

  it('sets user to null on 401', async () => {
    const { ApiError } = await import('@/api/client')
    vi.mocked(apiFetch).mockRejectedValue(new ApiError(401, 'Unauthorized'))
    await useAuthStore.getState().fetchCurrentUser()
    expect(useAuthStore.getState().user).toBeNull()
  })
})

describe('logout', () => {
  it('clears user and calls /auth/logout', async () => {
    useAuthStore.setState({ user: { id: 'u1', email: 'a@b.com', role: 'admin' } })
    vi.mocked(apiFetch).mockResolvedValue(undefined)
    await useAuthStore.getState().logout()
    expect(useAuthStore.getState().user).toBeNull()
    expect(apiFetch).toHaveBeenCalledWith('/api/v1/auth/logout', { method: 'POST' })
  })
})
```

Run: `npm test`

Expected: FAIL — `authStore` not found

- [ ] **Step 2: Create authStore.ts**

Create `services/frontend/src/features/auth/authStore.ts`:

```typescript
import { create } from 'zustand'
import { apiFetch, ApiError } from '@/api/client'

export interface User {
  id: string
  email: string
  role: 'admin' | 'contributor'
}

interface AuthState {
  user: User | null
  requestMagicLink: (email: string) => Promise<void>
  fetchCurrentUser: () => Promise<void>
  logout: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,

  requestMagicLink: async (email: string) => {
    await apiFetch('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email }),
    })
  },

  fetchCurrentUser: async () => {
    try {
      const user = await apiFetch<User>('/api/v1/users/me')
      set({ user })
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        set({ user: null })
      }
    }
  },

  logout: async () => {
    await apiFetch('/api/v1/auth/logout', { method: 'POST' })
    set({ user: null })
  },
}))
```

- [ ] **Step 3: Run tests**

```bash
cd services/frontend && npm test
```

Expected: 3 tests pass

- [ ] **Step 4: Commit**

```bash
git add services/frontend/src/features/auth/
git commit -m "feat: add auth store with fetchCurrentUser and logout"
```

---

## Task 4: Login and Verify pages

**Files:**
- Create: `services/frontend/src/features/auth/LoginPage.tsx`
- Create: `services/frontend/src/features/auth/VerifyPage.tsx`

- [ ] **Step 1: Create LoginPage.tsx**

Create `services/frontend/src/features/auth/LoginPage.tsx`:

```tsx
import { useState } from 'react'
import {
  Center,
  Stack,
  Title,
  Text,
  TextInput,
  Button,
  Alert,
  Paper,
} from '@mantine/core'
import { useAuthStore } from './authStore'

export function LoginPage() {
  const requestMagicLink = useAuthStore((s) => s.requestMagicLink)
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      await requestMagicLink(email)
      setSent(true)
    } catch {
      setError('Could not send magic link. Check the email address and try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Center h="100vh">
      <Paper shadow="md" p="xl" w={400}>
        <Stack>
          <Title order={2}>Family Photos</Title>
          {sent ? (
            <Alert color="green" title="Check your email">
              We sent a sign-in link to <strong>{email}</strong>. Click the link to continue.
            </Alert>
          ) : (
            <form onSubmit={handleSubmit}>
              <Stack>
                <Text c="dimmed" size="sm">
                  Enter your email to receive a sign-in link.
                </Text>
                {error && <Alert color="red">{error}</Alert>}
                <TextInput
                  label="Email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.currentTarget.value)}
                  required
                  autoFocus
                />
                <Button type="submit" loading={loading} fullWidth>
                  Send link
                </Button>
              </Stack>
            </form>
          )}
        </Stack>
      </Paper>
    </Center>
  )
}
```

- [ ] **Step 2: Create VerifyPage.tsx**

Create `services/frontend/src/features/auth/VerifyPage.tsx`:

```tsx
import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Center, Loader, Alert, Stack, Text } from '@mantine/core'
import { useAuthStore } from './authStore'
import { apiFetch, ApiError } from '@/api/client'

export function VerifyPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const fetchCurrentUser = useAuthStore((s) => s.fetchCurrentUser)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const token = params.get('token')
    if (!token) {
      setError('Missing token in URL.')
      return
    }
    apiFetch(`/api/v1/auth/verify?token=${encodeURIComponent(token)}`)
      .then(async () => {
        await fetchCurrentUser()
        navigate('/', { replace: true })
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError) {
          setError(err.detail)
        } else {
          setError('Verification failed. The link may have expired.')
        }
      })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  if (error) {
    return (
      <Center h="100vh">
        <Stack align="center">
          <Alert color="red" title="Sign-in failed">{error}</Alert>
          <Text
            component="a"
            href="/login"
            c="blue"
            size="sm"
          >
            Try again
          </Text>
        </Stack>
      </Center>
    )
  }

  return (
    <Center h="100vh">
      <Loader />
    </Center>
  )
}
```

- [ ] **Step 3: Verify TypeScript**

```bash
cd services/frontend && npm run build
```

Expected: build succeeds with no type errors

- [ ] **Step 4: Commit**

```bash
git add services/frontend/src/features/auth/LoginPage.tsx services/frontend/src/features/auth/VerifyPage.tsx
git commit -m "feat: add LoginPage and VerifyPage"
```

---

## Task 5: App shell, routing, and protected routes

**Files:**
- Create: `services/frontend/src/shared/ProtectedRoute.tsx`
- Create: `services/frontend/src/shared/AppShell.tsx`
- Modify: `services/frontend/src/App.tsx`

- [ ] **Step 1: Create ProtectedRoute.tsx**

Create `services/frontend/src/shared/ProtectedRoute.tsx`:

```tsx
import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/authStore'

export function ProtectedRoute() {
  const user = useAuthStore((s) => s.user)
  if (!user) return <Navigate to="/login" replace />
  return <Outlet />
}
```

- [ ] **Step 2: Create AppShell.tsx**

Create `services/frontend/src/shared/AppShell.tsx`:

```tsx
import { AppShell as MantineAppShell, Group, Text, Button } from '@mantine/core'
import { Outlet } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/authStore'

export function AppShell() {
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)

  return (
    <MantineAppShell
      header={{ height: 56 }}
      padding="md"
    >
      <MantineAppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Text fw={700} size="lg">Family Photos</Text>
          <Group>
            <Text size="sm" c="dimmed">{user?.email}</Text>
            <Button variant="subtle" size="sm" onClick={logout}>Sign out</Button>
          </Group>
        </Group>
      </MantineAppShell.Header>
      <MantineAppShell.Main>
        <Outlet />
      </MantineAppShell.Main>
    </MantineAppShell>
  )
}
```

- [ ] **Step 3: Wire up App.tsx with real routes and auth initialization**

Replace `services/frontend/src/App.tsx` with:

```tsx
import { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/authStore'
import { LoginPage } from '@/features/auth/LoginPage'
import { VerifyPage } from '@/features/auth/VerifyPage'
import { AppShell } from '@/shared/AppShell'
import { ProtectedRoute } from '@/shared/ProtectedRoute'

export function App() {
  const fetchCurrentUser = useAuthStore((s) => s.fetchCurrentUser)

  useEffect(() => {
    fetchCurrentUser()
  }, [fetchCurrentUser])

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/auth/verify" element={<VerifyPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route path="/" element={<div>Browse placeholder</div>} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
```

- [ ] **Step 4: Verify TypeScript**

```bash
cd services/frontend && npm run build
```

Expected: clean build

- [ ] **Step 5: Commit**

```bash
git add services/frontend/src/shared/ services/frontend/src/App.tsx
git commit -m "feat: add app shell, protected routes, and auth initialization"
```

---

## Task 6: ImageCard and ImageGrid components

**Files:**
- Create: `services/frontend/src/features/images/ImageCard.tsx`
- Create: `services/frontend/src/features/images/ImageGrid.tsx`

The backend `GET /api/v1/images` response shape (from the backend architecture doc):

```typescript
interface ImageListResponse {
  data: ImageSummary[]
  pagination: { total: number; limit: number; offset: number; hasMore: boolean }
}

interface ImageSummary {
  id: string
  imageId: string
  originalFilename: string
  thumbnailKey: string
  webKey: string
  originalKey: string
  width: number
  height: number
  tags: string[]
  people: string[]
  occasionCategory: string | null
  uploadedAt: string
}
```

Note: `thumbnailKey` is a storage key like `uploads/abc/thumbnail.webp`. For local development with `STORAGE_BACKEND=local`, the ingestion service does not expose files via HTTP — images will not render. Use `originalFilename` as fallback text. For production with S3, keys become pre-signed URLs.

- [ ] **Step 1: Define the shared image type**

Create `services/frontend/src/features/images/types.ts`:

```typescript
export interface ImageSummary {
  id: string
  imageId: string
  originalFilename: string
  thumbnailKey: string
  webKey: string
  originalKey: string
  width: number
  height: number
  tags: string[]
  people: string[]
  occasionCategory: string | null
  uploadedAt: string
}

export interface ImageListResponse {
  data: ImageSummary[]
  pagination: {
    total: number
    limit: number
    offset: number
    hasMore: boolean
  }
}
```

- [ ] **Step 2: Create ImageCard.tsx**

Create `services/frontend/src/features/images/ImageCard.tsx`:

```tsx
import { Card, Text, Badge, Group, Stack } from '@mantine/core'
import type { ImageSummary } from './types'

interface Props {
  image: ImageSummary
}

export function ImageCard({ image }: Props) {
  return (
    <Card shadow="sm" padding="sm" radius="md" withBorder>
      <Card.Section>
        <Stack
          h={180}
          align="center"
          justify="center"
          bg="gray.1"
          style={{ overflow: 'hidden' }}
        >
          <Text size="xs" c="dimmed" ta="center" px="xs">
            {image.originalFilename}
          </Text>
        </Stack>
      </Card.Section>
      <Stack mt="sm" gap="xs">
        <Text size="sm" fw={500} lineClamp={1}>
          {image.originalFilename}
        </Text>
        {image.tags.length > 0 && (
          <Group gap="xs">
            {image.tags.slice(0, 3).map((tag) => (
              <Badge key={tag} size="xs" variant="light">
                {tag}
              </Badge>
            ))}
            {image.tags.length > 3 && (
              <Badge size="xs" variant="outline">
                +{image.tags.length - 3}
              </Badge>
            )}
          </Group>
        )}
        {image.people.length > 0 && (
          <Text size="xs" c="dimmed">
            {image.people.join(', ')}
          </Text>
        )}
      </Stack>
    </Card>
  )
}
```

- [ ] **Step 3: Create ImageGrid.tsx**

Create `services/frontend/src/features/images/ImageGrid.tsx`:

```tsx
import { SimpleGrid, Text, Center, Loader } from '@mantine/core'
import { ImageCard } from './ImageCard'
import type { ImageSummary } from './types'

interface Props {
  images: ImageSummary[]
  loading: boolean
  error: string | null
}

export function ImageGrid({ images, loading, error }: Props) {
  if (loading) {
    return (
      <Center h={200}>
        <Loader />
      </Center>
    )
  }
  if (error) {
    return (
      <Center h={200}>
        <Text c="red">{error}</Text>
      </Center>
    )
  }
  if (images.length === 0) {
    return (
      <Center h={200}>
        <Text c="dimmed">No images found.</Text>
      </Center>
    )
  }
  return (
    <SimpleGrid cols={{ base: 1, sm: 2, md: 3, lg: 4 }} spacing="md">
      {images.map((img) => (
        <ImageCard key={img.id} image={img} />
      ))}
    </SimpleGrid>
  )
}
```

- [ ] **Step 4: Verify TypeScript**

```bash
cd services/frontend && npm run build
```

Expected: clean build

- [ ] **Step 5: Commit**

```bash
git add services/frontend/src/features/images/
git commit -m "feat: add ImageCard and ImageGrid components"
```

---

## Task 7: Filter store

**Files:**
- Create: `services/frontend/src/features/images/filterStore.ts`

The backend supports these query params on `GET /api/v1/images`:
- `tags` — comma-separated tag names
- `people` — person name
- `occasion` — occasion category string
- `limit` — integer (default 20)
- `offset` — integer

- [ ] **Step 1: Write failing test**

Create `services/frontend/src/features/images/filterStore.test.ts`:

```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { useFilterStore } from './filterStore'

beforeEach(() => {
  useFilterStore.setState({ tags: [], people: '', occasion: '', offset: 0 })
})

describe('useFilterStore', () => {
  it('builds empty query string when no filters set', () => {
    const qs = useFilterStore.getState().toQueryString()
    expect(qs).toBe('limit=20&offset=0')
  })

  it('includes tags when set', () => {
    useFilterStore.setState({ tags: ['beach', 'sunset'] })
    const qs = useFilterStore.getState().toQueryString()
    expect(qs).toContain('tags=beach%2Csunset')
  })

  it('includes people when set', () => {
    useFilterStore.setState({ people: 'Alice' })
    const qs = useFilterStore.getState().toQueryString()
    expect(qs).toContain('people=Alice')
  })

  it('includes occasion when set', () => {
    useFilterStore.setState({ occasion: 'birthday' })
    const qs = useFilterStore.getState().toQueryString()
    expect(qs).toContain('occasion=birthday')
  })

  it('resetFilters clears everything and resets offset', () => {
    useFilterStore.setState({ tags: ['x'], people: 'Bob', occasion: 'wedding', offset: 40 })
    useFilterStore.getState().resetFilters()
    const state = useFilterStore.getState()
    expect(state.tags).toEqual([])
    expect(state.people).toBe('')
    expect(state.occasion).toBe('')
    expect(state.offset).toBe(0)
  })
})
```

Run: `npm test`

Expected: FAIL — `filterStore` not found

- [ ] **Step 2: Create filterStore.ts**

Create `services/frontend/src/features/images/filterStore.ts`:

```typescript
import { create } from 'zustand'

const PAGE_SIZE = 20

interface FilterState {
  tags: string[]
  people: string
  occasion: string
  offset: number
  setTags: (tags: string[]) => void
  setPeople: (people: string) => void
  setOccasion: (occasion: string) => void
  nextPage: () => void
  prevPage: () => void
  resetFilters: () => void
  toQueryString: () => string
}

export const useFilterStore = create<FilterState>((set, get) => ({
  tags: [],
  people: '',
  occasion: '',
  offset: 0,

  setTags: (tags) => set({ tags, offset: 0 }),
  setPeople: (people) => set({ people, offset: 0 }),
  setOccasion: (occasion) => set({ occasion, offset: 0 }),
  nextPage: () => set((s) => ({ offset: s.offset + PAGE_SIZE })),
  prevPage: () => set((s) => ({ offset: Math.max(0, s.offset - PAGE_SIZE) })),
  resetFilters: () => set({ tags: [], people: '', occasion: '', offset: 0 }),

  toQueryString: () => {
    const { tags, people, occasion, offset } = get()
    const params = new URLSearchParams()
    if (tags.length > 0) params.set('tags', tags.join(','))
    if (people) params.set('people', people)
    if (occasion) params.set('occasion', occasion)
    params.set('limit', String(PAGE_SIZE))
    params.set('offset', String(offset))
    return params.toString()
  },
}))
```

- [ ] **Step 3: Run tests**

```bash
cd services/frontend && npm test
```

Expected: all tests pass (filter store + auth store + client)

- [ ] **Step 4: Commit**

```bash
git add services/frontend/src/features/images/filterStore.ts services/frontend/src/features/images/filterStore.test.ts
git commit -m "feat: add filter store with query string builder"
```

---

## Task 8: FilterBar component

**Files:**
- Create: `services/frontend/src/features/images/FilterBar.tsx`

- [ ] **Step 1: Create FilterBar.tsx**

Create `services/frontend/src/features/images/FilterBar.tsx`:

```tsx
import { Group, TagsInput, TextInput, Select, Button } from '@mantine/core'
import { useFilterStore } from './filterStore'

const OCCASION_OPTIONS = [
  { value: '', label: 'Any occasion' },
  { value: 'birthday', label: 'Birthday' },
  { value: 'wedding', label: 'Wedding' },
  { value: 'graduation', label: 'Graduation' },
  { value: 'holiday', label: 'Holiday' },
  { value: 'vacation', label: 'Vacation' },
  { value: 'work_event', label: 'Work event' },
  { value: 'party', label: 'Party' },
  { value: 'family_gathering', label: 'Family gathering' },
  { value: 'sports_event', label: 'Sports event' },
  { value: 'concert', label: 'Concert' },
  { value: 'conference', label: 'Conference' },
  { value: 'ceremony', label: 'Ceremony' },
  { value: 'casual', label: 'Casual' },
  { value: 'other', label: 'Other' },
]

export function FilterBar() {
  const tags = useFilterStore((s) => s.tags)
  const people = useFilterStore((s) => s.people)
  const occasion = useFilterStore((s) => s.occasion)
  const setTags = useFilterStore((s) => s.setTags)
  const setPeople = useFilterStore((s) => s.setPeople)
  const setOccasion = useFilterStore((s) => s.setOccasion)
  const resetFilters = useFilterStore((s) => s.resetFilters)

  const hasFilters = tags.length > 0 || people !== '' || occasion !== ''

  return (
    <Group align="flex-end" mb="md" wrap="wrap">
      <TagsInput
        label="Tags"
        placeholder="Add tag"
        value={tags}
        onChange={setTags}
        style={{ minWidth: 200 }}
      />
      <TextInput
        label="People"
        placeholder="Person name"
        value={people}
        onChange={(e) => setPeople(e.currentTarget.value)}
        style={{ minWidth: 160 }}
      />
      <Select
        label="Occasion"
        data={OCCASION_OPTIONS}
        value={occasion}
        onChange={(v) => setOccasion(v ?? '')}
        style={{ minWidth: 180 }}
      />
      {hasFilters && (
        <Button variant="subtle" color="gray" onClick={resetFilters}>
          Clear filters
        </Button>
      )}
    </Group>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
cd services/frontend && npm run build
```

Expected: clean build

- [ ] **Step 3: Commit**

```bash
git add services/frontend/src/features/images/FilterBar.tsx
git commit -m "feat: add FilterBar with tags, people, and occasion filters"
```

---

## Task 9: BrowsePage — wire everything together

**Files:**
- Create: `services/frontend/src/features/images/BrowsePage.tsx`
- Modify: `services/frontend/src/App.tsx`

- [ ] **Step 1: Create BrowsePage.tsx**

Create `services/frontend/src/features/images/BrowsePage.tsx`:

```tsx
import { useEffect, useState } from 'react'
import { Stack, Group, Button, Text } from '@mantine/core'
import { apiFetch } from '@/api/client'
import { useFilterStore } from './filterStore'
import { FilterBar } from './FilterBar'
import { ImageGrid } from './ImageGrid'
import type { ImageListResponse } from './types'

export function BrowsePage() {
  const toQueryString = useFilterStore((s) => s.toQueryString)
  const offset = useFilterStore((s) => s.offset)
  const nextPage = useFilterStore((s) => s.nextPage)
  const prevPage = useFilterStore((s) => s.prevPage)

  const [response, setResponse] = useState<ImageListResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    apiFetch<ImageListResponse>(`/api/v1/images?${toQueryString()}`)
      .then((data) => { if (!cancelled) { setResponse(data); setLoading(false) } })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load images')
          setLoading(false)
        }
      })
    return () => { cancelled = true }
  }, [toQueryString])

  const images = response?.data ?? []
  const pagination = response?.pagination

  return (
    <Stack>
      <FilterBar />
      <ImageGrid images={images} loading={loading} error={error} />
      {pagination && (
        <Group justify="space-between">
          <Text size="sm" c="dimmed">
            {pagination.total} image{pagination.total !== 1 ? 's' : ''}
          </Text>
          <Group>
            <Button
              variant="subtle"
              size="sm"
              disabled={offset === 0}
              onClick={prevPage}
            >
              Previous
            </Button>
            <Button
              variant="subtle"
              size="sm"
              disabled={!pagination.hasMore}
              onClick={nextPage}
            >
              Next
            </Button>
          </Group>
        </Group>
      )}
    </Stack>
  )
}
```

- [ ] **Step 2: Wire BrowsePage into App.tsx**

Replace the `Browse placeholder` div in `services/frontend/src/App.tsx`:

```tsx
import { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/authStore'
import { LoginPage } from '@/features/auth/LoginPage'
import { VerifyPage } from '@/features/auth/VerifyPage'
import { BrowsePage } from '@/features/images/BrowsePage'
import { AppShell } from '@/shared/AppShell'
import { ProtectedRoute } from '@/shared/ProtectedRoute'

export function App() {
  const fetchCurrentUser = useAuthStore((s) => s.fetchCurrentUser)

  useEffect(() => {
    fetchCurrentUser()
  }, [fetchCurrentUser])

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/auth/verify" element={<VerifyPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route path="/" element={<BrowsePage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
```

- [ ] **Step 3: Final build check**

```bash
cd services/frontend && npm run build
```

Expected: clean build — `dist/` populated, no TypeScript errors

- [ ] **Step 4: Run all tests**

```bash
cd services/frontend && npm test
```

Expected: all tests pass

- [ ] **Step 5: Commit**

```bash
git add services/frontend/src/features/images/BrowsePage.tsx services/frontend/src/App.tsx
git commit -m "feat: add BrowsePage, wire up browse route"
```

---

## Self-Review

### Spec coverage

| Requirement | Task |
|---|---|
| Vite + React + TypeScript scaffold | Task 1 |
| Mantine v7 with all CSS imports | Task 1 |
| API client: credentials:include, RFC7807 errors | Task 2 |
| Auth store: fetchCurrentUser, requestMagicLink, logout | Task 3 |
| GET /api/v1/users/me backend endpoint | Task 0 |
| Login page: email form → magic link | Task 4 |
| Verify page: token → cookie → redirect | Task 4 |
| Protected routes → redirect to /login | Task 5 |
| App shell with user email and sign out | Task 5 |
| fetchCurrentUser on app mount | Task 5 |
| ImageCard, ImageGrid components | Task 6 |
| Filter store with query string builder | Task 7 |
| FilterBar: tags, people, occasion | Task 8 |
| BrowsePage: fetches images, paginates | Task 9 |
| Vite proxy: /api/v1/ingest → 8080, /api → 8081 | Task 1 |

### Placeholder scan

No TBDs, TODOs, or "similar to Task N" references. All code blocks are complete.

### Type consistency

- `ImageSummary` defined in `types.ts` (Task 6), used in `ImageCard`, `ImageGrid`, `BrowsePage`
- `User` defined in `authStore.ts`, consumed in `AppShell` and `ProtectedRoute`
- `apiFetch<T>` from `client.ts` used consistently with type parameters
- `useFilterStore` actions (`setTags`, `setPeople`, `setOccasion`) match FilterBar usage exactly

---
