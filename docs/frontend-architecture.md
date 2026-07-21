# Frontend Architecture

_Note:_ Technical decisions for the React frontend: build tool, component library, styling, state management, image optimisation, and form handling.

## Decision History

| Date | Decision | Reason |
|------|----------|--------|
| Initial | Vite + React + TypeScript + Tailwind CSS | Fast builds, utility-first styling |
| 2026-04-25 | **Replaced Tailwind with Mantine** | Mantine provides TagsInput, Dropzone, DatePicker, Modal out of the box — all needed for this app. Eliminates 200+ lines of custom component code for inputs the design doesn't need to be custom. |

---

## Build Tool: Vite

**Decision:** Vite + React + TypeScript

**Rationale:**
- Native ES module HMR — dev server starts in milliseconds
- Rollup-based production builds with tree-shaking
- Zero-config TypeScript via esbuild
- PWA plugin available for progressive web app support

Alternatives rejected: CRA (deprecated, slow), Next.js (SSR/SSG unnecessary for a private SPA with no SEO requirements).

```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8081', // backend default port
        changeOrigin: true,
      },
    },
  },
  build: {
    sourcemap: true,
  },
})
```

---

## Component Library & Styling: Mantine

**Decision:** Mantine v7 (replaces Tailwind CSS)

**Rationale:**

Tailwind is utility-first and flexible, but this app needs specific complex components that Tailwind leaves to you to build from scratch:

| UI need | Tailwind | Mantine |
|---------|----------|---------|
| Tag multi-select with suggestions | ~80 lines custom | `<TagsInput>` |
| File dropzone with preview | react-dropzone + custom | `@mantine/dropzone` |
| Date range picker | third-party lib + styling | `@mantine/dates` |
| Modal / Drawer | hand-rolled | `<Modal>` / `<Drawer>` |
| Upload progress notifications | hand-rolled | `@mantine/notifications` |
| Form inputs (accessible, labelled) | manual | `<TextInput>`, `<Select>`, etc. |

Mantine v7 uses CSS variables — no runtime CSS-in-JS overhead.

### Packages

```bash
npm install @mantine/core @mantine/hooks @mantine/form @mantine/dropzone @mantine/dates @mantine/notifications dayjs
```

### Setup

```tsx
// src/main.tsx
import '@mantine/core/styles.css'
import '@mantine/dates/styles.css'
import '@mantine/dropzone/styles.css'
import '@mantine/notifications/styles.css'
import { MantineProvider } from '@mantine/core'
import { Notifications } from '@mantine/notifications'
import { ModalsProvider } from '@mantine/modals'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <MantineProvider>
      <Notifications />
      <ModalsProvider>
        <App />
      </ModalsProvider>
    </MantineProvider>
  </React.StrictMode>
)
```

### Theme

```typescript
// src/theme.ts
import { createTheme } from '@mantine/core'

export const theme = createTheme({
  primaryColor: 'blue',
  fontFamily: 'Inter, sans-serif',
  defaultRadius: 'md',
})
```

### Example: ImageCard with Mantine

```tsx
// components/ImageCard.tsx
import { Card, Badge, Group, Image, Overlay } from '@mantine/core'
import { useState } from 'react'

interface ImageCardProps {
  imageId: string
  thumbnailUrl: string
  alt: string
  tags: string[]
  onSelect: (id: string) => void
  selected: boolean
}

export const ImageCard = ({ imageId, thumbnailUrl, alt, tags, onSelect, selected }: ImageCardProps) => {
  const [hovered, setHovered] = useState(false)

  return (
    <Card
      shadow="sm"
      radius="md"
      withBorder={selected}
      style={{ cursor: 'pointer', outline: selected ? '3px solid var(--mantine-color-blue-5)' : undefined }}
      onClick={() => onSelect(imageId)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <Card.Section style={{ position: 'relative' }}>
        <Image src={thumbnailUrl} alt={alt} height={200} />
        {hovered && (
          <Overlay color="#000" backgroundOpacity={0.5} radius="md">
            <Group p="xs" style={{ position: 'absolute', bottom: 8, left: 8, flexWrap: 'wrap' }}>
              {tags.map((tag) => (
                <Badge key={tag} size="sm" variant="filled">{tag}</Badge>
              ))}
            </Group>
          </Overlay>
        )}
      </Card.Section>
    </Card>
  )
}
```

**Documentation:** https://mantine.dev/

---

## Authentication

**Not OAuth.** This app uses **magic link (passwordless email)** authentication only.

### How it works

```
1. User enters email → POST /api/v1/auth/login
   Backend: generates a one-time token, emails a magic link

2. User clicks the link → GET /api/v1/auth/verify?token=xxx
   Backend: validates token, sets an httpOnly JWT cookie, returns user info

3. All subsequent API calls include the cookie automatically (same-origin)

4. Logout → POST /api/v1/auth/logout
   Backend: clears the cookie
```

**The frontend never sees the JWT.** It is stored in an httpOnly cookie, which means:
- JavaScript cannot read or manipulate it
- It is sent automatically with every same-origin request
- No need to manage tokens in localStorage or Zustand

### What the frontend needs to track

The frontend only needs to know **who the user is** (for display and role-based UI). This requires a `/api/v1/users/me` endpoint on the backend that returns the current user from the cookie session.

> **Backend gap:** `GET /api/v1/users/me` does not exist yet. It needs to be added — reads the JWT from the cookie, returns the current user's id, email, name, and role.

### Auth store

```typescript
// stores/authStore.ts
import { create } from 'zustand'

interface User {
  id: string
  email: string
  name: string
  role: 'admin' | 'contributor'
}

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  requestMagicLink: (email: string) => Promise<void>
  fetchCurrentUser: () => Promise<void>
  logout: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,

  // Step 1: request magic link — backend sends email, no token returned
  requestMagicLink: async (email: string) => {
    const res = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    })
    if (!res.ok) throw new Error('Failed to send magic link')
    // No token to store — backend handles everything
  },

  // Step 2: after redirect from magic link, fetch user identity
  // The backend has already set the httpOnly cookie via /auth/verify
  fetchCurrentUser: async () => {
    const res = await fetch('/api/v1/users/me', { credentials: 'include' })
    if (!res.ok) {
      set({ user: null, isAuthenticated: false })
      return
    }
    const user = await res.json()
    set({ user, isAuthenticated: true })
  },

  logout: async () => {
    await fetch('/api/v1/auth/logout', { method: 'POST', credentials: 'include' })
    set({ user: null, isAuthenticated: false })
  },
}))
```

### Auth flow in the UI

```tsx
// pages/Login.tsx
import { useState } from 'react'
import { TextInput, Button, Stack, Text, Paper, Title } from '@mantine/core'
import { useAuthStore } from '@/stores/authStore'

export const Login = () => {
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)
  const [loading, setLoading] = useState(false)
  const requestMagicLink = useAuthStore((s) => s.requestMagicLink)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      await requestMagicLink(email)
      setSent(true)
    } finally {
      setLoading(false)
    }
  }

  if (sent) {
    return (
      <Paper p="xl" withBorder>
        <Title order={2}>Check your email</Title>
        <Text mt="sm" c="dimmed">
          We sent a login link to <strong>{email}</strong>. Click it to sign in.
        </Text>
      </Paper>
    )
  }

  return (
    <Paper p="xl" withBorder>
      <Title order={2} mb="md">Sign in</Title>
      <form onSubmit={handleSubmit}>
        <Stack>
          <TextInput
            label="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <Button type="submit" loading={loading}>Send magic link</Button>
        </Stack>
      </form>
    </Paper>
  )
}
```

```tsx
// pages/AuthVerify.tsx — rendered at /auth/verify?token=xxx
import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Center, Loader, Text } from '@mantine/core'
import { useAuthStore } from '@/stores/authStore'

export const AuthVerify = () => {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const fetchCurrentUser = useAuthStore((s) => s.fetchCurrentUser)

  useEffect(() => {
    const token = params.get('token')
    if (!token) { navigate('/login'); return }

    // Backend sets the cookie; we just need to fetch user identity
    fetch(`/api/v1/auth/verify?token=${token}`, { credentials: 'include' })
      .then((res) => {
        if (!res.ok) throw new Error('Invalid token')
        return fetchCurrentUser()
      })
      .then(() => navigate('/'))
      .catch(() => navigate('/login?error=invalid_token'))
  }, [])

  return (
    <Center h="100vh">
      <Stack align="center">
        <Loader />
        <Text>Signing you in…</Text>
      </Stack>
    </Center>
  )
}
```

---

## State Management: Zustand

**Decision:** Zustand (unchanged)

Small bundle (1KB gzipped), no provider boilerplate, fine-grained subscriptions. Fits the scope: auth state, image filters, upload queue.

### Filter store

```typescript
// stores/filterStore.ts
import { create } from 'zustand'

interface FilterState {
  tags: string[]
  people: string[]
  dateFrom: Date | null
  dateTo: Date | null
  occasionCategory: string | null
  setTags: (tags: string[]) => void
  setPeople: (people: string[]) => void
  setDateRange: (from: Date | null, to: Date | null) => void
  setOccasionCategory: (cat: string | null) => void
  clearFilters: () => void
}

export const useFilterStore = create<FilterState>((set) => ({
  tags: [],
  people: [],
  dateFrom: null,
  dateTo: null,
  occasionCategory: null,
  setTags: (tags) => set({ tags }),
  setPeople: (people) => set({ people }),
  setDateRange: (from, to) => set({ dateFrom: from, dateTo: to }),
  setOccasionCategory: (cat) => set({ occasionCategory: cat }),
  clearFilters: () => set({ tags: [], people: [], dateFrom: null, dateTo: null, occasionCategory: null }),
}))
```

### Upload queue store

```typescript
// stores/uploadStore.ts
import { create } from 'zustand'
import { notifications } from '@mantine/notifications'

interface UploadItem {
  id: string
  file: File
  preview: string
  tags: string[]
  people: string[]
  status: 'pending' | 'uploading' | 'completed' | 'failed'
  progress: number
  error?: string
}

interface UploadState {
  items: UploadItem[]
  addItems: (files: File[]) => void
  updateItem: (id: string, updates: Partial<UploadItem>) => void
  removeItem: (id: string) => void
  uploadAll: () => Promise<void>
  clearCompleted: () => void
}

export const useUploadStore = create<UploadState>((set, get) => ({
  items: [],

  addItems: (files) => {
    const newItems = files.map((file) => ({
      id: crypto.randomUUID(),
      file,
      preview: URL.createObjectURL(file),
      tags: [],
      people: [],
      status: 'pending' as const,
      progress: 0,
    }))
    set((state) => ({ items: [...state.items, ...newItems] }))
  },

  updateItem: (id, updates) =>
    set((state) => ({
      items: state.items.map((item) => (item.id === id ? { ...item, ...updates } : item)),
    })),

  removeItem: (id) =>
    set((state) => ({ items: state.items.filter((item) => item.id !== id) })),

  uploadAll: async () => {
    const pending = get().items.filter((item) => item.status === 'pending')

    for (const item of pending) {
      get().updateItem(item.id, { status: 'uploading' })

      const formData = new FormData()
      formData.append('image', item.file)

      // POST to ingestion service (handles resize/S3/metadata pipeline)
      // Ingestion service calls back to backend to register the image record
      try {
        const res = await fetch('/ingest/upload', {
          method: 'POST',
          body: formData,
          credentials: 'include',
        })
        if (!res.ok) throw new Error('Upload failed')

        const { imageId } = await res.json()

        // Apply tags and people to the registered image
        await fetch(`/api/v1/images/${imageId}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ tags: item.tags, people: item.people }),
        })

        get().updateItem(item.id, { status: 'completed', progress: 100 })
        notifications.show({ title: 'Uploaded', message: item.file.name, color: 'green' })
      } catch (err) {
        const error = err instanceof Error ? err.message : 'Unknown error'
        get().updateItem(item.id, { status: 'failed', error })
        notifications.show({ title: 'Upload failed', message: item.file.name, color: 'red' })
      }
    }
  },

  clearCompleted: () =>
    set((state) => ({ items: state.items.filter((item) => item.status !== 'completed') })),
}))
```

---

## Forms: React Hook Form + Zod + Mantine

React Hook Form integrates with Mantine inputs via the `Controller` wrapper or Mantine's own `useForm`. For complex forms (upload metadata), use React Hook Form + Zod for schema validation. For simpler forms, Mantine's `useForm` is fine.

### Upload form with Mantine components

```tsx
// components/UploadForm.tsx
import { useForm, zodResolver } from '@mantine/form'
import { TextInput, Select, Button, Stack, Group } from '@mantine/core'
import { TagsInput } from '@mantine/core'
import { DatePickerInput } from '@mantine/dates'
import { z } from 'zod'

const schema = z.object({
  tags: z.array(z.string()).min(1, 'Add at least one tag'),
  people: z.array(z.string()),
  occasionCategory: z.string().optional(),
  occasionName: z.string().optional(),
})

type UploadFormValues = z.infer<typeof schema>

const OCCASION_OPTIONS = [
  { value: 'birthday', label: 'Birthday' },
  { value: 'wedding', label: 'Wedding' },
  { value: 'graduation', label: 'Graduation' },
  { value: 'holiday', label: 'Holiday' },
  { value: 'vacation', label: 'Vacation' },
  { value: 'work_event', label: 'Work event' },
  { value: 'family_gathering', label: 'Family gathering' },
  { value: 'casual', label: 'Casual' },
  { value: 'other', label: 'Other' },
]

export const UploadMetadataForm = ({ onSubmit }: { onSubmit: (values: UploadFormValues) => void }) => {
  const form = useForm<UploadFormValues>({
    validate: zodResolver(schema),
    initialValues: { tags: [], people: [], occasionCategory: undefined, occasionName: undefined },
  })

  return (
    <form onSubmit={form.onSubmit(onSubmit)}>
      <Stack>
        <TagsInput
          label="Tags"
          placeholder="Type and press Enter"
          {...form.getInputProps('tags')}
        />
        <TagsInput
          label="People in these photos"
          placeholder="Add names"
          {...form.getInputProps('people')}
        />
        <Select
          label="Occasion"
          placeholder="Select occasion"
          data={OCCASION_OPTIONS}
          clearable
          {...form.getInputProps('occasionCategory')}
        />
        {form.values.occasionCategory === 'other' && (
          <TextInput
            label="Event name"
            {...form.getInputProps('occasionName')}
          />
        )}
        <Button type="submit">Apply & Upload</Button>
      </Stack>
    </form>
  )
}
```

### Dropzone for file selection

```tsx
// components/ImageDropzone.tsx
import { Dropzone, IMAGE_MIME_TYPE } from '@mantine/dropzone'
import { Group, Text, rem } from '@mantine/core'
import { IconUpload, IconPhoto, IconX } from '@tabler/icons-react'

interface ImageDropzoneProps {
  onDrop: (files: File[]) => void
}

export const ImageDropzone = ({ onDrop }: ImageDropzoneProps) => (
  <Dropzone
    onDrop={onDrop}
    accept={IMAGE_MIME_TYPE}
    maxSize={15 * 1024 * 1024} // 15MB
    multiple
  >
    <Group justify="center" gap="xl" mih={180} style={{ pointerEvents: 'none' }}>
      <Dropzone.Accept>
        <IconUpload size={52} color="var(--mantine-color-blue-6)" stroke={1.5} />
      </Dropzone.Accept>
      <Dropzone.Reject>
        <IconX size={52} color="var(--mantine-color-red-6)" stroke={1.5} />
      </Dropzone.Reject>
      <Dropzone.Idle>
        <IconPhoto size={52} color="var(--mantine-color-dimmed)" stroke={1.5} />
      </Dropzone.Idle>
      <div>
        <Text size="xl" inline>Drag images here or click to select</Text>
        <Text size="sm" c="dimmed" inline mt={7}>JPEG, PNG, HEIC — up to 15MB each</Text>
      </div>
    </Group>
  </Dropzone>
)
```

---

## Filter Bar

```tsx
// components/FilterBar.tsx
import { Group, TagsInput, Select, Button } from '@mantine/core'
import { DatePickerInput } from '@mantine/dates'
import { useFilterStore } from '@/stores/filterStore'

const OCCASION_OPTIONS = [
  { value: 'birthday', label: 'Birthday' },
  { value: 'wedding', label: 'Wedding' },
  // ... full list
]

export const FilterBar = () => {
  const { tags, people, dateFrom, dateTo, occasionCategory, setTags, setPeople, setDateRange, setOccasionCategory, clearFilters } =
    useFilterStore()

  return (
    <Group align="flex-end" wrap="wrap">
      <TagsInput
        label="Tags"
        placeholder="Filter by tag"
        value={tags}
        onChange={setTags}
      />
      <TagsInput
        label="People"
        placeholder="Filter by person"
        value={people}
        onChange={setPeople}
      />
      <DatePickerInput
        type="range"
        label="Date range"
        placeholder="Pick dates"
        value={[dateFrom, dateTo]}
        onChange={([from, to]) => setDateRange(from, to)}
        clearable
      />
      <Select
        label="Occasion"
        placeholder="Any"
        data={OCCASION_OPTIONS}
        value={occasionCategory}
        onChange={setOccasionCategory}
        clearable
      />
      <Button variant="subtle" onClick={clearFilters}>Clear</Button>
    </Group>
  )
}
```

---

## Image Optimisation & Lazy Loading

Browser-native lazy loading is now sufficient for most cases. Use `loading="lazy"` on images and the Intersection Observer API for custom behaviour.

### Lazy image component

```tsx
// components/LazyImage.tsx
import { useRef, useEffect, useState } from 'react'
import { Skeleton } from '@mantine/core'

interface LazyImageProps {
  thumbnailSrc: string
  fullSrc?: string
  alt: string
  height?: number
}

export const LazyImage = ({ thumbnailSrc, fullSrc, alt, height = 200 }: LazyImageProps) => {
  const [loaded, setLoaded] = useState(false)
  const [src, setSrc] = useState(thumbnailSrc)

  return (
    <>
      {!loaded && <Skeleton height={height} />}
      <img
        src={src}
        alt={alt}
        loading="lazy"
        style={{ display: loaded ? 'block' : 'none', width: '100%', height, objectFit: 'cover' }}
        onLoad={() => {
          setLoaded(true)
          if (fullSrc && src === thumbnailSrc) setSrc(fullSrc)
        }}
      />
    </>
  )
}
```

### Virtualised grid for large collections

Use `react-window` for galleries with 100+ images to avoid rendering off-screen DOM nodes.

```tsx
// components/VirtualImageGrid.tsx
import { FixedSizeGrid as Grid } from 'react-window'
import AutoSizer from 'react-virtualized-auto-sizer'
import { ImageCard } from './ImageCard'

const COLUMN_COUNT = 4
const CELL_SIZE = 260

export const VirtualImageGrid = ({ images }: { images: { id: string; thumbnailUrl: string; alt: string; tags: string[] }[] }) => {
  const rowCount = Math.ceil(images.length / COLUMN_COUNT)

  const Cell = ({ columnIndex, rowIndex, style }: { columnIndex: number; rowIndex: number; style: React.CSSProperties }) => {
    const index = rowIndex * COLUMN_COUNT + columnIndex
    if (index >= images.length) return null
    const img = images[index]
    return (
      <div style={{ ...style, padding: 8 }}>
        <ImageCard {...img} onSelect={() => {}} selected={false} />
      </div>
    )
  }

  return (
    <AutoSizer>
      {({ height, width }) => (
        <Grid columnCount={COLUMN_COUNT} columnWidth={CELL_SIZE} height={height} rowCount={rowCount} rowHeight={CELL_SIZE} width={width}>
          {Cell}
        </Grid>
      )}
    </AutoSizer>
  )
}
```

---

## Routing

React Router v6 with lazy-loaded routes.

```tsx
// App.tsx
import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { LoadingOverlay } from '@mantine/core'
import { useAuthStore } from '@/stores/authStore'

const Browse = lazy(() => import('./pages/Browse'))
const Upload = lazy(() => import('./pages/Upload'))
const ImageDetail = lazy(() => import('./pages/ImageDetail'))
const Login = lazy(() => import('./pages/Login'))
const AuthVerify = lazy(() => import('./pages/AuthVerify'))
const Admin = lazy(() => import('./pages/Admin'))

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />
}

const AdminRoute = ({ children }: { children: React.ReactNode }) => {
  const user = useAuthStore((s) => s.user)
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== 'admin') return <Navigate to="/" replace />
  return <>{children}</>
}

export const App = () => (
  <BrowserRouter>
    <Suspense fallback={<LoadingOverlay visible />}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/auth/verify" element={<AuthVerify />} />
        <Route path="/" element={<ProtectedRoute><Browse /></ProtectedRoute>} />
        <Route path="/upload" element={<ProtectedRoute><Upload /></ProtectedRoute>} />
        <Route path="/images/:id" element={<ProtectedRoute><ImageDetail /></ProtectedRoute>} />
        <Route path="/admin" element={<AdminRoute><Admin /></AdminRoute>} />
      </Routes>
    </Suspense>
  </BrowserRouter>
)
```

---

## Testing

### Stack

| Layer | Tool |
|-------|------|
| Unit + component | Vitest + React Testing Library |
| E2E | Playwright |
| Component dev | Storybook (optional, useful for the image grid) |

### Vitest setup

```bash
npm install -D vitest @vitest/ui @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
```

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
})
```

```typescript
// src/test/setup.ts
import '@testing-library/jest-dom'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'
afterEach(cleanup)
```

### Example component test

```typescript
// components/ImageCard.test.tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MantineProvider } from '@mantine/core'
import { ImageCard } from './ImageCard'

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <MantineProvider>{children}</MantineProvider>
)

describe('ImageCard', () => {
  it('calls onSelect when clicked', async () => {
    const onSelect = vi.fn()
    render(
      <ImageCard imageId="img_1" thumbnailUrl="/test.jpg" alt="test" tags={['beach']} onSelect={onSelect} selected={false} />,
      { wrapper }
    )
    await userEvent.click(screen.getByAltText('test'))
    expect(onSelect).toHaveBeenCalledWith('img_1')
  })
})
```

### Playwright config

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  use: { baseURL: 'http://localhost:3000', trace: 'on-first-retry' },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'mobile', use: { ...devices['Pixel 5'] } },
  ],
  webServer: { command: 'npm run dev', url: 'http://localhost:3000', reuseExistingServer: !process.env.CI },
})
```

---

## Project Structure

Feature-based layout — keep components, hooks, and stores co-located by feature.

```
src/
├── features/
│   ├── auth/
│   │   ├── pages/
│   │   │   ├── Login.tsx
│   │   │   └── AuthVerify.tsx
│   │   └── stores/
│   │       └── authStore.ts
│   ├── images/
│   │   ├── components/
│   │   │   ├── ImageCard.tsx
│   │   │   ├── ImageGrid.tsx
│   │   │   ├── VirtualImageGrid.tsx
│   │   │   └── FilterBar.tsx
│   │   ├── pages/
│   │   │   ├── Browse.tsx
│   │   │   └── ImageDetail.tsx
│   │   └── stores/
│   │       └── filterStore.ts
│   ├── upload/
│   │   ├── components/
│   │   │   ├── ImageDropzone.tsx
│   │   │   └── UploadMetadataForm.tsx
│   │   ├── pages/
│   │   │   └── Upload.tsx
│   │   └── stores/
│   │       └── uploadStore.ts
│   └── admin/
│       ├── pages/
│       │   └── Admin.tsx
│       └── components/
│           └── UserTable.tsx
├── shared/
│   ├── components/
│   │   ├── LazyImage.tsx
│   │   └── ErrorBoundary.tsx
│   └── hooks/
│       └── useApi.ts
├── App.tsx
└── main.tsx
```

---

## TypeScript Configuration

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "jsx": "react-jsx",
    "module": "ESNext",
    "target": "ESNext",
    "lib": ["ESNext", "DOM", "DOM.Iterable"],
    "moduleResolution": "bundler",
    "paths": { "@/*": ["./src/*"] }
  }
}
```

---

## Known Backend Gaps

Items the frontend needs that aren't yet in the backend:

| Gap | Required for |
|-----|-------------|
| `GET /api/v1/users/me` | Fetching current user identity after login; auth store |
| Pagination total count in `GET /api/v1/images` | Showing "150 photos" in the browse header; next/prev pagination |

---

## Summary

| Concern | Decision |
|---------|----------|
| Build tool | Vite + React + TypeScript |
| Component library | Mantine v7 |
| State management | Zustand |
| Forms | Mantine `useForm` + Zod |
| File upload UI | `@mantine/dropzone` |
| Date picking | `@mantine/dates` |
| Auth | Magic link (passwordless) — httpOnly cookie, no client-side token |
| Image virtualisation | react-window (100+ images) |
| Unit/component tests | Vitest + React Testing Library |
| E2E tests | Playwright |
