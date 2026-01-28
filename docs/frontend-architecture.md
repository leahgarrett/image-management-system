# Frontend Architecture

_Note:_ This document outlines the technical decisions for the React frontend, including build tool selection, styling approach, state management, image optimization strategies, and form handling for uploads and tagging.

## React Setup: Build Tool Selection

### Recommended Approach: Vite

**Rationale:**
- **Lightning-fast HMR:** Instant hot module replacement using native ES modules
- **Optimized builds:** Rollup-based production bundling with tree-shaking
- **Modern by default:** Built for ES2015+, no legacy baggage
- **TypeScript support:** Zero-config TypeScript with fast transpilation via esbuild
- **Simple configuration:** Minimal setup compared to webpack-based tools
- **Development experience:** Server start in milliseconds, not seconds

### Alternatives Considered

| Tool | Pros | Cons |
|------|------|------|
| **Vite** | Instant HMR, fast builds, modern, simple config, excellent DX | Smaller ecosystem than webpack, newer tool |
| **Create React App** | Zero config, battle-tested, large community | Slow dev server, slow builds, webpack complexity, deprecated |
| **Next.js** | SSR/SSG built-in, file-based routing, API routes, image optimization | Overkill for SPA, opinionated structure, heavier bundle |

**Decision:** Vite provides the best development experience for a single-page application with fast iteration cycles. Next.js is unnecessary since we don't need SEO or SSR for a private family photo app, and CRA is being deprecated in favor of modern alternatives.

**Example Configuration:**

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
      '@components': path.resolve(__dirname, './src/components'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@utils': path.resolve(__dirname, './src/utils'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'ui-vendor': ['react-dropzone', 'react-hook-form'],
        },
      },
    },
  },
})
```

**Documentation:** https://vitejs.dev/

---

## CSS/Styling Approach

### Recommended Approach: Tailwind CSS

**Rationale:**
- **Utility-first:** Rapid UI development without context switching
- **Consistent design system:** Predefined spacing, colors, typography scales
- **Responsive by default:** Mobile-first breakpoint system
- **Production optimization:** PurgeCSS removes unused styles automatically
- **No CSS-in-JS runtime:** Zero JavaScript overhead
- **Team productivity:** Designers and developers use same vocabulary

### Alternatives Considered

| Approach | Pros | Cons |
|----------|------|------|
| **Tailwind CSS** | Fast development, consistent design, no naming conflicts, small bundle | Verbose HTML, learning curve |
| **styled-components** | Component-scoped styles, dynamic styling, TypeScript support | Runtime overhead, larger bundle, SSR complexity |
| **CSS Modules** | Scoped styles, familiar CSS syntax, no runtime cost | Manual naming, no design system, boilerplate |

**Decision:** Tailwind CSS provides the best balance of development speed, consistency, and performance for a utility-focused UI. The design system constraints help maintain visual coherence across the app.

**Example Component:**

```tsx
// components/ImageCard.tsx
import { useState } from 'react'
import { LazyLoadImage } from 'react-lazy-load-image-component'

interface ImageCardProps {
  imageId: string
  thumbnailUrl: string
  alt: string
  tags: string[]
  onSelect: (id: string) => void
  selected: boolean
}

export const ImageCard: React.FC<ImageCardProps> = ({
  imageId,
  thumbnailUrl,
  alt,
  tags,
  onSelect,
  selected,
}) => {
  const [isHovered, setIsHovered] = useState(false)

  return (
    <div
      className={`
        relative overflow-hidden rounded-lg cursor-pointer
        transition-all duration-200 ease-in-out
        ${selected ? 'ring-4 ring-blue-500' : 'hover:shadow-lg'}
      `}
      onClick={() => onSelect(imageId)}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <LazyLoadImage
        src={thumbnailUrl}
        alt={alt}
        className="w-full h-full object-cover"
        effect="blur"
      />
      
      {isHovered && (
        <div className="absolute inset-0 bg-black bg-opacity-50 flex items-end p-4">
          <div className="flex flex-wrap gap-2">
            {tags.map((tag) => (
              <span
                key={tag}
                className="px-2 py-1 bg-blue-500 text-white text-xs rounded-full"
              >
                {tag}
              </span>
            ))}
          </div>
        </div>
      )}
      
      {selected && (
        <div className="absolute top-2 right-2">
          <div className="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          </div>
        </div>
      )}
    </div>
  )
}
```

**Tailwind Configuration:**

```javascript
// tailwind.config.js
/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0f9ff',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
        },
      },
      spacing: {
        72: '18rem',
        84: '21rem',
        96: '24rem',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/aspect-ratio'),
  ],
}
```

**Documentation:** https://tailwindcss.com/

---

## State Management

### Recommended Approach: Zustand

**Rationale:**
- **Minimal boilerplate:** Simple API, less code than Redux or Context
- **TypeScript-first:** Excellent type inference
- **Performance:** Fine-grained subscriptions, no unnecessary re-renders
- **DevTools support:** Redux DevTools integration
- **Small bundle:** 1KB gzipped
- **No provider wrapping:** Direct store access

### Alternatives Considered

| Solution | Pros | Cons |
|----------|------|------|
| **Zustand** | Simple API, small bundle, great performance, TypeScript support | Newer, smaller ecosystem |
| **Context API** | Built-in, no dependencies, simple for basic cases | Performance issues with frequent updates, verbose |
| **Redux** | Battle-tested, large ecosystem, powerful middleware | Boilerplate-heavy, overkill for this project, larger bundle |

**Decision:** Zustand provides the right balance of simplicity and power. Our state management needs (auth, filters, upload queue) don't require Redux's complexity, and Zustand solves Context API's performance issues.

### State Structure

**Auth Store:**
```typescript
// stores/authStore.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id: string
  email: string
  name: string
  role: 'admin' | 'editor' | 'viewer'
  permissions: string[]
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  login: (email: string) => Promise<void>
  verifyToken: (token: string) => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      
      login: async (email: string) => {
        const response = await fetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email }),
        })
        if (!response.ok) throw new Error('Login failed')
        // Magic link sent
      },
      
      verifyToken: async (token: string) => {
        const response = await fetch(`/api/v1/auth/verify?token=${token}`)
        const data = await response.json()
        
        set({
          user: data.user,
          token: data.jwtToken,
          isAuthenticated: true,
        })
      },
      
      logout: () => {
        set({ user: null, token: null, isAuthenticated: false })
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ token: state.token }), // Only persist token
    }
  )
)
```

**Filter Store:**
```typescript
// stores/filterStore.ts
import { create } from 'zustand'

interface FilterState {
  tags: string[]
  people: string[]
  dateFrom: Date | null
  dateTo: Date | null
  searchQuery: string
  
  setTags: (tags: string[]) => void
  setPeople: (people: string[]) => void
  setDateRange: (from: Date | null, to: Date | null) => void
  setSearchQuery: (query: string) => void
  clearFilters: () => void
}

export const useFilterStore = create<FilterState>((set) => ({
  tags: [],
  people: [],
  dateFrom: null,
  dateTo: null,
  searchQuery: '',
  
  setTags: (tags) => set({ tags }),
  setPeople: (people) => set({ people }),
  setDateRange: (from, to) => set({ dateFrom: from, dateTo: to }),
  setSearchQuery: (query) => set({ searchQuery: query }),
  clearFilters: () => set({
    tags: [],
    people: [],
    dateFrom: null,
    dateTo: null,
    searchQuery: '',
  }),
}))
```

**Upload Queue Store:**
```typescript
// stores/uploadStore.ts
import { create } from 'zustand'

interface UploadItem {
  id: string
  file: File
  preview: string
  tags: string[]
  status: 'pending' | 'uploading' | 'processing' | 'completed' | 'failed'
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
      status: 'pending' as const,
      progress: 0,
    }))
    set((state) => ({ items: [...state.items, ...newItems] }))
  },
  
  updateItem: (id, updates) => {
    set((state) => ({
      items: state.items.map((item) =>
        item.id === id ? { ...item, ...updates } : item
      ),
    }))
  },
  
  removeItem: (id) => {
    set((state) => ({
      items: state.items.filter((item) => item.id !== id),
    }))
  },
  
  uploadAll: async () => {
    const items = get().items.filter((item) => item.status === 'pending')
    
    for (const item of items) {
      try {
        get().updateItem(item.id, { status: 'uploading' })
        
        const formData = new FormData()
        formData.append('image', item.file)
        formData.append('tags', item.tags.join(','))
        
        const response = await fetch('/api/v1/ingest/upload', {
          method: 'POST',
          body: formData,
          headers: {
            Authorization: `Bearer ${useAuthStore.getState().token}`,
          },
        })
        
        if (!response.ok) throw new Error('Upload failed')
        
        get().updateItem(item.id, { status: 'completed', progress: 100 })
      } catch (error) {
        get().updateItem(item.id, {
          status: 'failed',
          error: error instanceof Error ? error.message : 'Unknown error',
        })
      }
    }
  },
  
  clearCompleted: () => {
    set((state) => ({
      items: state.items.filter((item) => item.status !== 'completed'),
    }))
  },
}))
```

**Documentation:** https://github.com/pmndrs/zustand

---

## Image Optimization & Lazy Loading

### Recommended Libraries

#### 1. React Lazy Load Image Component

**Purpose:** Lazy load images as they enter viewport

```tsx
import { LazyLoadImage } from 'react-lazy-load-image-component'
import 'react-lazy-load-image-component/src/effects/blur.css'

export const OptimizedImage: React.FC<{
  src: string
  alt: string
  width?: number
  height?: number
}> = ({ src, alt, width, height }) => {
  return (
    <LazyLoadImage
      src={src}
      alt={alt}
      width={width}
      height={height}
      effect="blur"
      threshold={200} // Start loading 200px before entering viewport
      placeholderSrc="/placeholder.jpg" // Low-res placeholder
    />
  )
}
```

**Documentation:** https://www.npmjs.com/package/react-lazy-load-image-component

#### 2. React Window (Virtualization)

**Purpose:** Render only visible images in large collections

```tsx
import { FixedSizeGrid as Grid } from 'react-window'
import AutoSizer from 'react-virtualized-auto-sizer'

interface ImageGridProps {
  images: Array<{ id: string; thumbnailUrl: string; alt: string }>
}

export const VirtualizedImageGrid: React.FC<ImageGridProps> = ({ images }) => {
  const COLUMN_COUNT = 4
  const CELL_SIZE = 250
  const ROW_COUNT = Math.ceil(images.length / COLUMN_COUNT)

  const Cell = ({ columnIndex, rowIndex, style }: any) => {
    const index = rowIndex * COLUMN_COUNT + columnIndex
    if (index >= images.length) return null

    const image = images[index]

    return (
      <div style={style} className="p-2">
        <ImageCard {...image} />
      </div>
    )
  }

  return (
    <AutoSizer>
      {({ height, width }) => (
        <Grid
          columnCount={COLUMN_COUNT}
          columnWidth={CELL_SIZE}
          height={height}
          rowCount={ROW_COUNT}
          rowHeight={CELL_SIZE}
          width={width}
        >
          {Cell}
        </Grid>
      )}
    </AutoSizer>
  )
}
```

**Documentation:** https://react-window.vercel.app/

#### 3. Progressive Image Loading

**Custom Hook:**
```tsx
// hooks/useProgressiveImage.ts
import { useEffect, useState } from 'react'

export const useProgressiveImage = (lowQualitySrc: string, highQualitySrc: string) => {
  const [src, setSrc] = useState(lowQualitySrc)

  useEffect(() => {
    setSrc(lowQualitySrc)

    const img = new Image()
    img.src = highQualitySrc

    img.onload = () => {
      setSrc(highQualitySrc)
    }
  }, [lowQualitySrc, highQualitySrc])

  return src
}

// Usage
export const ProgressiveImage: React.FC<{
  lowQualitySrc: string
  highQualitySrc: string
  alt: string
}> = ({ lowQualitySrc, highQualitySrc, alt }) => {
  const src = useProgressiveImage(lowQualitySrc, highQualitySrc)
  const isLoaded = src === highQualitySrc

  return (
    <img
      src={src}
      alt={alt}
      className={`transition-all duration-300 ${
        isLoaded ? 'blur-0' : 'blur-sm'
      }`}
    />
  )
}
```

### Performance Optimization Strategies

**Image Format Selection:**
```tsx
// utils/imageUrl.ts
export const getOptimalImageUrl = (
  imageId: string,
  variant: 'thumbnail' | 'web' | 'original'
): string => {
  const supportsWebP = document
    .createElement('canvas')
    .toDataURL('image/webp')
    .indexOf('data:image/webp') === 0

  const format = supportsWebP ? 'webp' : 'jpg'
  return `https://cdn.example.com/${imageId}/${variant}.${format}`
}
```

**Intersection Observer for Custom Lazy Loading:**
```tsx
// hooks/useIntersectionObserver.ts
import { useEffect, useRef, useState } from 'react'

export const useIntersectionObserver = (options?: IntersectionObserverInit) => {
  const [isVisible, setIsVisible] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting) {
        setIsVisible(true)
        observer.disconnect()
      }
    }, options)

    if (ref.current) {
      observer.observe(ref.current)
    }

    return () => observer.disconnect()
  }, [options])

  return { ref, isVisible }
}

// Usage
export const LazyImageCard: React.FC<{ imageUrl: string }> = ({ imageUrl }) => {
  const { ref, isVisible } = useIntersectionObserver({ threshold: 0.1 })

  return (
    <div ref={ref} className="min-h-[200px]">
      {isVisible ? (
        <img src={imageUrl} alt="Lazy loaded" />
      ) : (
        <div className="bg-gray-200 animate-pulse w-full h-full" />
      )}
    </div>
  )
}
```

---

## Form Handling

### Recommended Approach: React Hook Form + Zod

**Rationale:**
- **Performance:** Uncontrolled inputs reduce re-renders
- **TypeScript integration:** Type-safe form data with Zod schemas
- **Validation:** Declarative schema-based validation
- **Developer experience:** Minimal boilerplate, intuitive API
- **Bundle size:** 9KB gzipped

### Upload Form Example

```tsx
// components/UploadForm.tsx
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useDropzone } from 'react-dropzone'
import { useUploadStore } from '@/stores/uploadStore'

const uploadSchema = z.object({
  tags: z.array(z.string()).min(1, 'Add at least one tag'),
  occasion: z.enum([
    'birthday',
    'wedding',
    'graduation',
    'holiday',
    'vacation',
    'casual',
    'other',
  ]),
  eventName: z.string().optional(),
  people: z.array(z.string()),
})

type UploadFormData = z.infer<typeof uploadSchema>

export const UploadForm: React.FC = () => {
  const addItems = useUploadStore((state) => state.addItems)
  
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<UploadFormData>({
    resolver: zodResolver(uploadSchema),
    defaultValues: {
      tags: [],
      people: [],
    },
  })

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    accept: {
      'image/jpeg': ['.jpg', '.jpeg'],
      'image/png': ['.png'],
      'image/heic': ['.heic'],
    },
    maxSize: 15 * 1024 * 1024, // 15MB
    multiple: true,
    onDrop: (acceptedFiles) => {
      addItems(acceptedFiles)
    },
  })

  const onSubmit = (data: UploadFormData) => {
    console.log('Form data:', data)
    // Apply metadata to all items in upload queue
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Dropzone */}
      <div
        {...getRootProps()}
        className={`
          border-2 border-dashed rounded-lg p-12 text-center cursor-pointer
          transition-colors duration-200
          ${isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300'}
        `}
      >
        <input {...getInputProps()} />
        <div className="space-y-2">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            stroke="currentColor"
            fill="none"
            viewBox="0 0 48 48"
          >
            <path
              d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
              strokeWidth={2}
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <p className="text-lg font-medium text-gray-700">
            {isDragActive
              ? 'Drop files here'
              : 'Drag & drop images, or click to select'}
          </p>
          <p className="text-sm text-gray-500">
            JPEG, PNG, HEIC up to 15MB each
          </p>
        </div>
      </div>

      {/* Tags Input */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Tags *
        </label>
        <TagInput
          value={watch('tags')}
          onChange={(tags) => setValue('tags', tags)}
        />
        {errors.tags && (
          <p className="mt-1 text-sm text-red-600">{errors.tags.message}</p>
        )}
      </div>

      {/* Occasion Select */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Occasion *
        </label>
        <select
          {...register('occasion')}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
        >
          <option value="">Select occasion</option>
          <option value="birthday">Birthday</option>
          <option value="wedding">Wedding</option>
          <option value="graduation">Graduation</option>
          <option value="holiday">Holiday</option>
          <option value="vacation">Vacation</option>
          <option value="casual">Casual</option>
          <option value="other">Other</option>
        </select>
        {errors.occasion && (
          <p className="mt-1 text-sm text-red-600">{errors.occasion.message}</p>
        )}
      </div>

      {/* Event Name (conditional) */}
      {watch('occasion') === 'other' && (
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Event Name
          </label>
          <input
            {...register('eventName')}
            type="text"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
            placeholder="Enter event name"
          />
        </div>
      )}

      {/* People Input */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          People
        </label>
        <PeopleInput
          value={watch('people')}
          onChange={(people) => setValue('people', people)}
        />
      </div>

      {/* Submit Button */}
      <button
        type="submit"
        className="w-full px-4 py-2 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        Apply Tags & Upload
      </button>
    </form>
  )
}
```

### Tag Input Component (Multi-value)

```tsx
// components/TagInput.tsx
import { useState } from 'react'

interface TagInputProps {
  value: string[]
  onChange: (tags: string[]) => void
  suggestions?: string[]
}

export const TagInput: React.FC<TagInputProps> = ({
  value,
  onChange,
  suggestions = [],
}) => {
  const [input, setInput] = useState('')
  const [showSuggestions, setShowSuggestions] = useState(false)

  const filteredSuggestions = suggestions.filter(
    (tag) =>
      tag.toLowerCase().includes(input.toLowerCase()) &&
      !value.includes(tag)
  )

  const addTag = (tag: string) => {
    if (tag && !value.includes(tag)) {
      onChange([...value, tag])
      setInput('')
    }
  }

  const removeTag = (tagToRemove: string) => {
    onChange(value.filter((tag) => tag !== tagToRemove))
  }

  return (
    <div className="relative">
      <div className="flex flex-wrap gap-2 p-2 border border-gray-300 rounded-md min-h-[42px]">
        {value.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-sm"
          >
            {tag}
            <button
              type="button"
              onClick={() => removeTag(tag)}
              className="hover:text-blue-600"
            >
              ×
            </button>
          </span>
        ))}
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onFocus={() => setShowSuggestions(true)}
          onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              addTag(input.trim())
            }
          }}
          placeholder={value.length === 0 ? 'Type and press Enter' : ''}
          className="flex-1 min-w-[120px] outline-none"
        />
      </div>

      {/* Suggestions Dropdown */}
      {showSuggestions && filteredSuggestions.length > 0 && (
        <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto">
          {filteredSuggestions.map((tag) => (
            <button
              key={tag}
              type="button"
              onClick={() => addTag(tag)}
              className="w-full px-3 py-2 text-left hover:bg-gray-100"
            >
              {tag}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
```

**Documentation:**
- React Hook Form: https://react-hook-form.com/
- Zod: https://zod.dev/
- React Dropzone: https://react-dropzone.js.org/

---

## Batch Tagging Interface

```tsx
// components/BatchTaggingModal.tsx
import { useForm } from 'react-hook-form'
import { useUploadStore } from '@/stores/uploadStore'

export const BatchTaggingModal: React.FC<{ isOpen: boolean; onClose: () => void }> = ({
  isOpen,
  onClose,
}) => {
  const items = useUploadStore((state) => state.items)
  const updateItem = useUploadStore((state) => state.updateItem)
  
  const { register, handleSubmit, watch, setValue } = useForm<{
    commonTags: string[]
    applyToAll: boolean
  }>({
    defaultValues: {
      commonTags: [],
      applyToAll: true,
    },
  })

  const onSubmit = (data: { commonTags: string[]; applyToAll: boolean }) => {
    if (data.applyToAll) {
      items.forEach((item) => {
        updateItem(item.id, {
          tags: [...new Set([...item.tags, ...data.commonTags])],
        })
      })
    }
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="bg-white rounded-lg p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <h2 className="text-2xl font-bold mb-4">Batch Tag Images</h2>
        
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Add Tags to All Images
            </label>
            <TagInput
              value={watch('commonTags')}
              onChange={(tags) => setValue('commonTags', tags)}
            />
          </div>

          {/* Image Preview Grid with Individual Tag Editing */}
          <div className="grid grid-cols-3 gap-4 max-h-96 overflow-y-auto">
            {items.map((item) => (
              <div key={item.id} className="space-y-2">
                <img
                  src={item.preview}
                  alt="Preview"
                  className="w-full h-32 object-cover rounded"
                />
                <TagInput
                  value={item.tags}
                  onChange={(tags) => updateItem(item.id, { tags })}
                />
              </div>
            ))}
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              Apply Tags
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
```

---

## Complete Project Structure

```
src/
├── components/
│   ├── ImageCard.tsx
│   ├── ImageGrid.tsx
│   ├── VirtualizedImageGrid.tsx
│   ├── UploadForm.tsx
│   ├── TagInput.tsx
│   ├── PeopleInput.tsx
│   ├── BatchTaggingModal.tsx
│   ├── FilterBar.tsx
│   └── Layout.tsx
├── hooks/
│   ├── useProgressiveImage.ts
│   ├── useIntersectionObserver.ts
│   ├── useImageFetch.ts
│   └── useAuth.ts
├── stores/
│   ├── authStore.ts
│   ├── filterStore.ts
│   └── uploadStore.ts
├── utils/
│   ├── imageUrl.ts
│   ├── api.ts
│   └── validation.ts
├── pages/
│   ├── Browse.tsx
│   ├── Upload.tsx
│   ├── ImageDetail.tsx
│   ├── Login.tsx
│   └── Admin.tsx
├── App.tsx
├── main.tsx
└── index.css
```

---

## Performance Best Practices

1. **Code Splitting:** Dynamic imports for routes
```tsx
import { lazy, Suspense } from 'react'

const Browse = lazy(() => import('./pages/Browse'))
const Upload = lazy(() => import('./pages/Upload'))

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/" element={<Browse />} />
        <Route path="/upload" element={<Upload />} />
      </Routes>
    </Suspense>
  )
}
```

2. **Image Loading Priority:**
   - Above-the-fold images: Eager loading
   - Below-the-fold: Lazy loading with intersection observer
   - Thumbnails: Load on scroll
   - High-res: Load on demand (detail view)

3. **Memoization:**
```tsx
import { memo } from 'react'

export const ImageCard = memo<ImageCardProps>(({ imageId, ...props }) => {
  // Component logic
}, (prevProps, nextProps) => {
  // Custom comparison
  return prevProps.imageId === nextProps.imageId &&
         prevProps.selected === nextProps.selected
})
```

4. **Virtual Scrolling:** Use `react-window` for 100+ images

---

## Progressive Web App (PWA) Implementation

### Overview

Converting the application to a PWA provides a mobile-first experience without requiring native app development. Users can install the app on their home screen, access it offline, and use device features like the camera for photo uploads.

### PWA Plugin Configuration

**Install Vite PWA Plugin:**
```bash
npm install vite-plugin-pwa workbox-window -D
```

**Vite Configuration:**
```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'masked-icon.svg'],
      manifest: {
        name: 'Family Photo Manager',
        short_name: 'Photos',
        description: 'Private family photo collection and management',
        theme_color: '#0ea5e9',
        background_color: '#ffffff',
        display: 'standalone',
        orientation: 'any',
        scope: '/',
        start_url: '/',
        icons: [
          {
            src: 'pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,webp}'],
        runtimeCaching: [
          {
            // Cache thumbnail images
            urlPattern: /^https:\/\/cdn\.example\.com\/.*\/thumbnail\.(webp|jpg)$/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'image-thumbnails',
              expiration: {
                maxEntries: 500,
                maxAgeSeconds: 60 * 60 * 24 * 30, // 30 days
              },
              cacheableResponse: {
                statuses: [0, 200],
              },
            },
          },
          {
            // Network-first for API calls
            urlPattern: /^https:\/\/api\.example\.com\/api\/v1\/.*/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              networkTimeoutSeconds: 10,
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 60 * 5, // 5 minutes
              },
            },
          },
          {
            // Stale-while-revalidate for web-optimized images
            urlPattern: /^https:\/\/cdn\.example\.com\/.*\/web\.(webp|jpg)$/,
            handler: 'StaleWhileRevalidate',
            options: {
              cacheName: 'image-web-optimized',
              expiration: {
                maxEntries: 100,
                maxAgeSeconds: 60 * 60 * 24 * 7, // 7 days
              },
            },
          },
        ],
        cleanupOutdatedCaches: true,
        skipWaiting: true,
        clientsClaim: true,
      },
      devOptions: {
        enabled: true, // Enable in dev for testing
        type: 'module',
      },
    }),
  ],
})
```

### Service Worker Registration

```tsx
// src/main.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'
import { registerSW } from 'virtual:pwa-register'

// Register service worker with update prompt
const updateSW = registerSW({
  onNeedRefresh() {
    if (confirm('New content available. Reload?')) {
      updateSW(true)
    }
  },
  onOfflineReady() {
    console.log('App ready to work offline')
  },
})

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
```

### Camera Access for Mobile Upload

**HTML File Input with Camera Capture:**
```tsx
// components/MobileCameraUpload.tsx
import { useRef } from 'react'
import { useUploadStore } from '@/stores/uploadStore'

export const MobileCameraUpload: React.FC = () => {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const addItems = useUploadStore((state) => state.addItems)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      addItems(Array.from(e.target.files))
    }
  }

  return (
    <div className="space-y-4">
      {/* Take Photo Button */}
      <button
        onClick={() => fileInputRef.current?.click()}
        className="w-full py-4 bg-blue-600 text-white rounded-lg flex items-center justify-center gap-2"
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
          />
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
          />
        </svg>
        Take Photo
      </button>

      {/* Hidden file inputs */}
      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png,image/heic"
        capture="environment" // Use back camera
        multiple
        onChange={handleFileChange}
        className="hidden"
      />
      
      {/* Choose from Library */}
      <input
        type="file"
        accept="image/jpeg,image/png,image/heic"
        multiple
        onChange={handleFileChange}
        className="hidden"
        id="library-upload"
      />
      <label
        htmlFor="library-upload"
        className="block w-full py-4 bg-gray-100 text-gray-700 rounded-lg text-center cursor-pointer"
      >
        Choose from Library
      </label>
    </div>
  )
}
```

### Offline Support with IndexedDB

**Store pending uploads when offline:**
```typescript
// utils/offlineStorage.ts
import { openDB, DBSchema, IDBPDatabase } from 'idb'

interface OfflineDB extends DBSchema {
  'pending-uploads': {
    key: string
    value: {
      id: string
      file: Blob
      fileName: string
      tags: string[]
      metadata: Record<string, any>
      timestamp: number
    }
  }
  'cached-images': {
    key: string
    value: {
      imageId: string
      thumbnailBlob: Blob
      metadata: Record<string, any>
      cachedAt: number
    }
  }
}

let db: IDBPDatabase<OfflineDB> | null = null

export const initOfflineDB = async () => {
  if (db) return db
  
  db = await openDB<OfflineDB>('photo-manager-offline', 1, {
    upgrade(db) {
      db.createObjectStore('pending-uploads', { keyPath: 'id' })
      db.createObjectStore('cached-images', { keyPath: 'imageId' })
    },
  })
  
  return db
}

export const savePendingUpload = async (upload: {
  id: string
  file: File
  tags: string[]
  metadata: Record<string, any>
}) => {
  const database = await initOfflineDB()
  
  await database.put('pending-uploads', {
    id: upload.id,
    file: upload.file,
    fileName: upload.file.name,
    tags: upload.tags,
    metadata: upload.metadata,
    timestamp: Date.now(),
  })
}

export const getPendingUploads = async () => {
  const database = await initOfflineDB()
  return database.getAll('pending-uploads')
}

export const removePendingUpload = async (id: string) => {
  const database = await initOfflineDB()
  await database.delete('pending-uploads', id)
}

export const syncPendingUploads = async () => {
  const pendingUploads = await getPendingUploads()
  
  for (const upload of pendingUploads) {
    try {
      const formData = new FormData()
      formData.append('image', upload.file, upload.fileName)
      formData.append('tags', upload.tags.join(','))
      
      const response = await fetch('/api/v1/ingest/upload', {
        method: 'POST',
        body: formData,
      })
      
      if (response.ok) {
        await removePendingUpload(upload.id)
      }
    } catch (error) {
      console.error('Failed to sync upload:', error)
      // Keep in queue for next sync
    }
  }
}
```

**Online/Offline Detection:**
```tsx
// hooks/useOnlineStatus.ts
import { useState, useEffect } from 'react'
import { syncPendingUploads } from '@/utils/offlineStorage'

export const useOnlineStatus = () => {
  const [isOnline, setIsOnline] = useState(navigator.onLine)

  useEffect(() => {
    const handleOnline = async () => {
      setIsOnline(true)
      // Sync pending uploads when back online
      await syncPendingUploads()
    }

    const handleOffline = () => {
      setIsOnline(false)
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  return isOnline
}

// Usage in component
export const OfflineIndicator: React.FC = () => {
  const isOnline = useOnlineStatus()

  if (isOnline) return null

  return (
    <div className="fixed top-0 left-0 right-0 bg-yellow-500 text-white py-2 px-4 text-center z-50">
      <p>You're offline. Uploads will sync when connection is restored.</p>
    </div>
  )
}
```

### Install Prompt

```tsx
// components/InstallPrompt.tsx
import { useState, useEffect } from 'react'

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

export const InstallPrompt: React.FC = () => {
  const [installPrompt, setInstallPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [isInstalled, setIsInstalled] = useState(false)

  useEffect(() => {
    const handler = (e: Event) => {
      e.preventDefault()
      setInstallPrompt(e as BeforeInstallPromptEvent)
    }

    window.addEventListener('beforeinstallprompt', handler)

    // Check if already installed
    if (window.matchMedia('(display-mode: standalone)').matches) {
      setIsInstalled(true)
    }

    return () => window.removeEventListener('beforeinstallprompt', handler)
  }, [])

  const handleInstall = async () => {
    if (!installPrompt) return

    await installPrompt.prompt()
    const { outcome } = await installPrompt.userChoice

    if (outcome === 'accepted') {
      setIsInstalled(true)
      setInstallPrompt(null)
    }
  }

  if (isInstalled || !installPrompt) return null

  return (
    <div className="fixed bottom-4 left-4 right-4 bg-white rounded-lg shadow-lg p-4 border border-gray-200 z-50 md:left-auto md:w-96">
      <h3 className="font-semibold text-lg mb-2">Install Photo Manager</h3>
      <p className="text-gray-600 text-sm mb-4">
        Install this app on your device for quick access and offline support.
      </p>
      <div className="flex gap-2">
        <button
          onClick={handleInstall}
          className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Install
        </button>
        <button
          onClick={() => setInstallPrompt(null)}
          className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
        >
          Not Now
        </button>
      </div>
    </div>
  )
}
```

### Mobile-Optimized Upload Flow

```tsx
// components/MobileUploadFlow.tsx
import { useState } from 'react'
import { MobileCameraUpload } from './MobileCameraUpload'
import { useUploadStore } from '@/stores/uploadStore'
import { useOnlineStatus } from '@/hooks/useOnlineStatus'
import { savePendingUpload } from '@/utils/offlineStorage'

export const MobileUploadFlow: React.FC = () => {
  const items = useUploadStore((state) => state.items)
  const uploadAll = useUploadStore((state) => state.uploadAll)
  const isOnline = useOnlineStatus()
  const [step, setStep] = useState<'capture' | 'tag' | 'confirm'>('capture')

  const handleUpload = async () => {
    if (isOnline) {
      await uploadAll()
    } else {
      // Save to IndexedDB for later sync
      for (const item of items) {
        await savePendingUpload({
          id: item.id,
          file: item.file,
          tags: item.tags,
          metadata: {},
        })
      }
      alert('Photos saved. They will upload when you\'re back online.')
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* Step indicator */}
      <div className="bg-white border-b px-4 py-3">
        <div className="flex justify-between items-center max-w-md mx-auto">
          <div className={`flex-1 text-center ${step === 'capture' ? 'text-blue-600' : 'text-gray-400'}`}>
            1. Capture
          </div>
          <div className={`flex-1 text-center ${step === 'tag' ? 'text-blue-600' : 'text-gray-400'}`}>
            2. Tag
          </div>
          <div className={`flex-1 text-center ${step === 'confirm' ? 'text-blue-600' : 'text-gray-400'}`}>
            3. Confirm
          </div>
        </div>
      </div>

      {/* Step content */}
      <div className="max-w-md mx-auto px-4 py-6">
        {step === 'capture' && (
          <>
            <MobileCameraUpload />
            {items.length > 0 && (
              <button
                onClick={() => setStep('tag')}
                className="w-full mt-4 py-3 bg-blue-600 text-white rounded-lg"
              >
                Continue ({items.length} photos)
              </button>
            )}
          </>
        )}

        {step === 'tag' && (
          <>
            {/* Tag interface */}
            <button
              onClick={() => setStep('confirm')}
              className="w-full mt-4 py-3 bg-blue-600 text-white rounded-lg"
            >
              Review & Upload
            </button>
          </>
        )}

        {step === 'confirm' && (
          <>
            {/* Confirmation */}
            <button
              onClick={handleUpload}
              className="w-full mt-4 py-3 bg-green-600 text-white rounded-lg"
            >
              {isOnline ? 'Upload Now' : 'Save for Later (Offline)'}
            </button>
          </>
        )}
      </div>
    </div>
  )
}
```

### Testing PWA Features

**Local HTTPS for testing:**
```bash
# Install mkcert for local SSL
brew install mkcert
mkcert -install
mkcert localhost 127.0.0.1 ::1

# Update vite.config.ts
server: {
  https: {
    key: fs.readFileSync('./localhost-key.pem'),
    cert: fs.readFileSync('./localhost.pem'),
  },
  port: 3000,
}
```

**Chrome DevTools PWA Testing:**
1. Open DevTools → Application tab
2. Check Manifest, Service Workers, Storage
3. Use Lighthouse for PWA audit
4. Test offline mode in Network tab

**Documentation:**
- Vite PWA: https://vite-pwa-org.netlify.app/
- Workbox: https://developer.chrome.com/docs/workbox/
- PWA Best Practices: https://web.dev/progressive-web-apps/

---

## Testing Strategy

### Overview

A comprehensive testing strategy ensures code quality and maintainability for a small team. The approach prioritizes **ease of use, fast feedback, and minimal configuration** over exhaustive coverage, focusing on testing user-facing behavior rather than implementation details.

### Recommended Test Stack

#### 1. Vitest (Unit & Integration Tests)

**Why Vitest over Jest:**
- **Native Vite integration:** Zero additional configuration
- **Faster execution:** 2-10x faster than Jest
- **ESM support:** Works with ES modules out of the box
- **Compatible API:** Drop-in replacement for Jest (same syntax)
- **TypeScript:** First-class TypeScript support
- **Watch mode:** Instant re-runs on file changes

**Installation:**
```bash
npm install -D vitest @vitest/ui @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
```

**Configuration:**
```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/mockData',
      ],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

**Test Setup:**
```typescript
// src/test/setup.ts
import { expect, afterEach } from 'vitest'
import { cleanup } from '@testing-library/react'
import * as matchers from '@testing-library/jest-dom/matchers'

expect.extend(matchers)

// Cleanup after each test
afterEach(() => {
  cleanup()
})

// Mock IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
  takeRecords() {
    return []
  }
}
```

**Documentation:** https://vitest.dev/

#### 2. React Testing Library (Component Tests)

**Philosophy:** Test components the way users interact with them, not implementation details.

**Example Component Test:**
```typescript
// components/ImageCard.test.tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ImageCard } from './ImageCard'

describe('ImageCard', () => {
  const defaultProps = {
    imageId: 'img_123',
    thumbnailUrl: 'https://example.com/thumb.jpg',
    alt: 'Family vacation photo',
    tags: ['vacation', 'beach', '2024'],
    onSelect: vi.fn(),
    selected: false,
  }

  it('renders image with correct alt text', () => {
    render(<ImageCard {...defaultProps} />)
    
    const image = screen.getByAltText('Family vacation photo')
    expect(image).toBeInTheDocument()
  })

  it('shows tags on hover', async () => {
    const user = userEvent.setup()
    render(<ImageCard {...defaultProps} />)
    
    const card = screen.getByRole('button')
    await user.hover(card)
    
    expect(screen.getByText('vacation')).toBeVisible()
    expect(screen.getByText('beach')).toBeVisible()
  })

  it('calls onSelect when clicked', async () => {
    const user = userEvent.setup()
    const onSelect = vi.fn()
    
    render(<ImageCard {...defaultProps} onSelect={onSelect} />)
    
    await user.click(screen.getByRole('button'))
    expect(onSelect).toHaveBeenCalledWith('img_123')
  })

  it('shows selected state with checkmark', () => {
    render(<ImageCard {...defaultProps} selected={true} />)
    
    const checkmark = screen.getByRole('img', { hidden: true })
    expect(checkmark).toBeInTheDocument()
  })
})
```

**Documentation:** https://testing-library.com/react

#### 3. Playwright (E2E Tests)

**Why Playwright over Cypress:**
- **Better performance:** Faster test execution
- **Multiple browsers:** Chrome, Firefox, Safari, Edge
- **True multi-tab support:** Test workflows across tabs
- **Network interception:** Built-in API mocking
- **TypeScript-first:** Excellent TypeScript support
- **Less flakiness:** Auto-waits for elements

**Installation:**
```bash
npm install -D @playwright/test
npx playwright install
```

**Configuration:**
```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
  ],

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
})
```

**Example E2E Test:**
```typescript
// e2e/upload-flow.spec.ts
import { test, expect } from '@playwright/test'
import path from 'path'

test.describe('Upload Flow', () => {
  test('user can upload and tag images', async ({ page }) => {
    await page.goto('/')
    
    // Login
    await page.getByRole('button', { name: 'Login' }).click()
    await page.getByPlaceholder('Email').fill('user@example.com')
    await page.getByRole('button', { name: 'Send Magic Link' }).click()
    
    // Navigate to upload
    await page.getByRole('link', { name: 'Upload' }).click()
    
    // Upload files
    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles([
      path.join(__dirname, 'fixtures/test-image-1.jpg'),
      path.join(__dirname, 'fixtures/test-image-2.jpg'),
    ])
    
    // Wait for previews to load
    await expect(page.getByText('2 photos')).toBeVisible()
    
    // Add tags
    await page.getByPlaceholder('Type and press Enter').fill('vacation')
    await page.keyboard.press('Enter')
    await page.getByPlaceholder('Type and press Enter').fill('beach')
    await page.keyboard.press('Enter')
    
    // Select occasion
    await page.selectOption('select[name="occasion"]', 'vacation')
    
    // Submit
    await page.getByRole('button', { name: 'Apply Tags & Upload' }).click()
    
    // Verify success
    await expect(page.getByText('Upload complete')).toBeVisible({ timeout: 10000 })
  })

  test('shows offline indicator when disconnected', async ({ page, context }) => {
    await page.goto('/')
    
    // Simulate offline
    await context.setOffline(true)
    
    await expect(page.getByText(/you're offline/i)).toBeVisible()
    
    // Go back online
    await context.setOffline(false)
    await expect(page.getByText(/you're offline/i)).not.toBeVisible()
  })
})
```

**Documentation:** https://playwright.dev/

#### 4. Storybook (Component Development & Visual Testing)

**Why Storybook:**
- **Component isolation:** Develop components in isolation
- **Living documentation:** Auto-generates component docs
- **Visual regression:** Catch UI bugs with snapshots
- **Accessibility testing:** Built-in a11y addon
- **Mobile testing:** Test responsive designs
- **Design handoff:** Share components with designers

**Installation:**
```bash
npx storybook@latest init
npm install -D @storybook/addon-a11y @storybook/addon-interactions
```

**Configuration:**
```typescript
// .storybook/main.ts
import type { StorybookConfig } from '@storybook/react-vite'

const config: StorybookConfig = {
  stories: ['../src/**/*.mdx', '../src/**/*.stories.@(js|jsx|ts|tsx)'],
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    '@storybook/addon-interactions',
    '@storybook/addon-a11y',
  ],
  framework: {
    name: '@storybook/react-vite',
    options: {},
  },
  docs: {
    autodocs: 'tag',
  },
}

export default config
```

**Example Story:**
```typescript
// components/ImageCard.stories.tsx
import type { Meta, StoryObj } from '@storybook/react'
import { fn } from '@storybook/test'
import { ImageCard } from './ImageCard'

const meta = {
  title: 'Components/ImageCard',
  component: ImageCard,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  args: {
    onSelect: fn(),
  },
} satisfies Meta<typeof ImageCard>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    imageId: 'img_123',
    thumbnailUrl: 'https://picsum.photos/300/200',
    alt: 'Sample image',
    tags: ['vacation', 'beach'],
    selected: false,
  },
}

export const Selected: Story = {
  args: {
    ...Default.args,
    selected: true,
  },
}

export const WithManyTags: Story = {
  args: {
    ...Default.args,
    tags: ['vacation', 'beach', 'summer', '2024', 'family', 'sunset'],
  },
}

export const Mobile: Story = {
  args: Default.args,
  parameters: {
    viewport: {
      defaultViewport: 'mobile1',
    },
  },
}
```

**Visual Regression Testing:**
```typescript
// stories/ImageCard.test.ts
import { test, expect } from '@storybook/test'

test('ImageCard renders correctly', async ({ canvasElement }) => {
  const canvas = within(canvasElement)
  
  await expect(canvas.getByAltText('Sample image')).toBeInTheDocument()
  await expect(canvas.getByText('vacation')).toBeInTheDocument()
})
```

**Documentation:** https://storybook.js.org/

### Testing Pyramid Strategy

```
        /\
       /  \    E2E (Playwright)
      /    \   5-10 critical user flows
     /------\  
    /        \ Integration (Vitest + RTL)
   /          \ 20-30 component interactions
  /------------\
 /              \ Unit (Vitest)
/________________\ 50-100 business logic, utils, hooks
```

**Coverage Targets:**
- Unit Tests: 80%+ coverage
- Integration Tests: Key user workflows
- E2E Tests: Critical paths only (login, upload, browse)

### Testing Best Practices for Small Teams

#### 1. Test User Behavior, Not Implementation

**❌ Bad (Testing Implementation):**
```typescript
it('increments counter state', () => {
  const { result } = renderHook(() => useState(0))
  act(() => result.current[1](result.current[0] + 1))
  expect(result.current[0]).toBe(1)
})
```

**✅ Good (Testing Behavior):**
```typescript
it('increments counter when button clicked', async () => {
  const user = userEvent.setup()
  render(<Counter />)
  
  await user.click(screen.getByRole('button', { name: 'Increment' }))
  expect(screen.getByText('Count: 1')).toBeInTheDocument()
})
```

#### 2. Use Testing Library Queries in Priority Order

1. **getByRole** - Most accessible and resilient
2. **getByLabelText** - Forms and inputs
3. **getByPlaceholderText** - Last resort for inputs
4. **getByText** - User-visible text
5. **getByTestId** - Only when nothing else works

#### 3. Mock External Dependencies

```typescript
// utils/api.test.ts
import { vi } from 'vitest'
import { fetchImages } from './api'

// Mock fetch globally
global.fetch = vi.fn()

describe('fetchImages', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('fetches images with filters', async () => {
    const mockResponse = {
      data: [{ id: '1', url: 'test.jpg' }],
      pagination: { total: 1 },
    }
    
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    })
    
    const result = await fetchImages({ tags: ['vacation'] })
    
    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/images'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ tags: ['vacation'] }),
      })
    )
    expect(result).toEqual(mockResponse)
  })
})
```

#### 4. Test Zustand Stores in Isolation

```typescript
// stores/authStore.test.ts
import { renderHook, act } from '@testing-library/react'
import { useAuthStore } from './authStore'

describe('authStore', () => {
  beforeEach(() => {
    // Reset store state
    useAuthStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
    })
  })

  it('updates authentication state after login', async () => {
    const { result } = renderHook(() => useAuthStore())
    
    await act(async () => {
      await result.current.verifyToken('test-token')
    })
    
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user).toBeDefined()
  })

  it('clears state on logout', () => {
    const { result } = renderHook(() => useAuthStore())
    
    // Set initial state
    act(() => {
      useAuthStore.setState({
        user: { id: '1', email: 'test@example.com' },
        token: 'abc123',
        isAuthenticated: true,
      })
    })
    
    act(() => {
      result.current.logout()
    })
    
    expect(result.current.user).toBeNull()
    expect(result.current.token).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
  })
})
```

#### 5. Test Custom Hooks

```typescript
// hooks/useIntersectionObserver.test.ts
import { renderHook } from '@testing-library/react'
import { useIntersectionObserver } from './useIntersectionObserver'

describe('useIntersectionObserver', () => {
  it('sets isVisible when element intersects', () => {
    const { result } = renderHook(() => useIntersectionObserver())
    
    expect(result.current.isVisible).toBe(false)
    
    // Simulate intersection
    const [entry] = mockIntersectionObserverEntry({ isIntersecting: true })
    const callback = global.IntersectionObserver.mock.calls[0][0]
    callback([entry])
    
    expect(result.current.isVisible).toBe(true)
  })
})
```

### Running Tests

**Package.json scripts:**
```json
{
  "scripts": {
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest --coverage",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "storybook": "storybook dev -p 6006",
    "build-storybook": "storybook build"
  }
}
```

**CI Pipeline (.github/workflows/test.yml):**
```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm ci
      - run: npm run test:coverage
      - uses: codecov/codecov-action@v3

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: npm ci
      - run: npx playwright install --with-deps
      - run: npm run test:e2e
      - uses: actions/upload-artifact@v3
        if: failure()
        with:
          name: playwright-report
          path: playwright-report/
```

---

## Web Components: Not Recommended

### Why Skip Web Components for This Project

**Decision:** Use **React components** with Storybook, not Web Components.

**Rationale:**

**Cons of Web Components:**
- **React friction:** Awkward integration with React's synthetic events
- **TypeScript complexity:** Harder to type-check custom elements
- **State management:** Can't use Zustand directly
- **Learning curve:** Additional concepts for team to learn
- **Tooling immaturity:** Less mature ecosystem compared to React
- **Bundle size:** Shadow DOM polyfills for older browsers
- **Developer experience:** Slower hot reload, less debugging support

**Pros of React Components (Current Approach):**
- ✅ **Team familiarity:** React is already the chosen framework
- ✅ **Ecosystem:** Vast library ecosystem (Tailwind, React Hook Form, etc.)
- ✅ **TypeScript:** First-class TypeScript support
- ✅ **Tooling:** Excellent DevTools, Fast Refresh, ESLint plugins
- ✅ **State management:** Seamless Zustand integration
- ✅ **Testing:** Mature testing libraries (RTL, Vitest)

**When Web Components Make Sense:**
- Building a design system used across multiple frameworks
- Creating embeddable widgets for third-party sites
- Building framework-agnostic component library

**For this project:** Stick with React components. Use Storybook for component isolation and documentation instead of Web Components.

---

## Maintainability Best Practices

### 1. Code Organization

**Follow Feature-Based Structure:**
```
src/
├── features/
│   ├── auth/
│   │   ├── components/
│   │   │   ├── LoginForm.tsx
│   │   │   ├── LoginForm.test.tsx
│   │   │   └── LoginForm.stories.tsx
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   └── useAuth.test.ts
│   │   ├── stores/
│   │   │   ├── authStore.ts
│   │   │   └── authStore.test.ts
│   │   └── utils/
│   │       ├── validation.ts
│   │       └── validation.test.ts
│   ├── images/
│   │   ├── components/
│   │   ├── hooks/
│   │   └── stores/
│   └── upload/
├── shared/
│   ├── components/
│   ├── hooks/
│   └── utils/
├── test/
│   ├── setup.ts
│   └── utils.tsx
└── App.tsx
```

### 2. TypeScript Configuration

**Strict Mode Enabled:**
```json
// tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "jsx": "react-jsx",
    "module": "ESNext",
    "target": "ESNext",
    "lib": ["ESNext", "DOM", "DOM.Iterable"],
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

### 3. ESLint & Prettier

**ESLint Configuration:**
```javascript
// eslint.config.js
import js from '@eslint/js'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import tseslint from '@typescript-eslint/eslint-plugin'

export default [
  js.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    plugins: {
      '@typescript-eslint': tseslint,
      'react': react,
      'react-hooks': reactHooks,
    },
    rules: {
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/explicit-function-return-type': 'off',
      'no-console': ['warn', { allow: ['warn', 'error'] }],
    },
  },
]
```

**Prettier Configuration:**
```json
// .prettierrc
{
  "semi": false,
  "singleQuote": true,
  "trailingComma": "es5",
  "printWidth": 90,
  "tabWidth": 2,
  "arrowParens": "always"
}
```

### 4. Pre-commit Hooks

**Husky + Lint-Staged:**
```bash
npm install -D husky lint-staged
npx husky init
```

```json
// package.json
{
  "lint-staged": {
    "*.{ts,tsx}": [
      "eslint --fix",
      "prettier --write",
      "vitest related --run"
    ]
  }
}
```

```bash
# .husky/pre-commit
npm run lint-staged
```

### 5. Component Documentation Standards

**Use JSDoc for Component Props:**
```typescript
interface ImageCardProps {
  /** Unique identifier for the image */
  imageId: string
  
  /** URL to the thumbnail image (300px) */
  thumbnailUrl: string
  
  /** Alt text for accessibility */
  alt: string
  
  /** Array of tag names associated with the image */
  tags: string[]
  
  /** Callback when image is selected/deselected */
  onSelect: (id: string) => void
  
  /** Whether the image is currently selected */
  selected: boolean
}
```

### 6. Error Boundaries

```tsx
// components/ErrorBoundary.tsx
import { Component, ErrorInfo, ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = {
    hasError: false,
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught:', error, errorInfo)
    // Send to error tracking service (e.g., Sentry)
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div className="p-8 text-center">
          <h2 className="text-xl font-bold text-red-600">Something went wrong</h2>
          <p className="text-gray-600 mt-2">{this.state.error?.message}</p>
          <button
            onClick={() => this.setState({ hasError: false })}
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded"
          >
            Try again
          </button>
        </div>
      )
    }

    return this.props.children
  }
}

// Usage
<ErrorBoundary>
  <ImageGrid />
</ErrorBoundary>
```

### 7. Performance Monitoring

```typescript
// utils/monitoring.ts
export const measurePerformance = (metricName: string) => {
  const start = performance.now()
  
  return () => {
    const duration = performance.now() - start
    console.log(`${metricName}: ${duration.toFixed(2)}ms`)
    
    // Send to analytics
    if (window.gtag) {
      window.gtag('event', 'timing_complete', {
        name: metricName,
        value: Math.round(duration),
      })
    }
  }
}

// Usage
const measure = measurePerformance('image-grid-render')
// ... render logic
measure()
```

---

## Summary

The frontend architecture prioritizes **performance, developer experience, and user experience** with modern tooling. **Vite** provides lightning-fast development builds, **Tailwind CSS** enables rapid UI development with design consistency, **Zustand** offers lightweight state management, and specialized libraries handle image optimization and form interactions efficiently. This stack balances modern best practices with practical simplicity, avoiding over-engineering while delivering excellent performance for browsing and managing large photo collections.

**Key Technologies:**
- **Build Tool:** Vite with React + TypeScript
- **Styling:** Tailwind CSS with custom design tokens
- **State Management:** Zustand for global state
- **Image Optimization:** react-lazy-load-image-component + react-window
- **Forms:** React Hook Form + Zod validation
- **File Upload:** react-dropzone

**Bundle Size Target:** < 200KB gzipped (initial load)
