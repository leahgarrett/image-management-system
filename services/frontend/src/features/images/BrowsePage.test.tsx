import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'
import { MemoryRouter } from 'react-router-dom'
import { useFilterStore } from './filterStore'
import { BrowsePage } from './BrowsePage'

vi.mock('@/api/client', () => ({
  apiFetch: vi.fn(),
  ApiError: class ApiError extends Error {
    constructor(public status: number, public detail: string) { super(detail) }
  },
}))

import { apiFetch } from '@/api/client'

const wrap = (ui: React.ReactElement) =>
  render(
    <MemoryRouter>
      <MantineProvider>{ui}</MantineProvider>
    </MemoryRouter>
  )

const emptyResponse = {
  data: [],
  pagination: { total: 0, limit: 20, offset: 0, hasMore: false },
}

const oneImageResponse = {
  data: [{
    id: '1', imageId: 'img-1', originalFilename: 'photo.webp',
    thumbnailKey: 't', webKey: 'w', originalKey: 'o',
    width: 100, height: 100, tags: [], people: [],
    occasionCategory: null, uploadedAt: '2024-01-01T00:00:00Z',
  }],
  pagination: { total: 1, limit: 20, offset: 0, hasMore: false },
}

beforeEach(() => {
  useFilterStore.setState({ tags: [], people: '', occasion: '', offset: 0 })
  vi.clearAllMocks()
})

describe('BrowsePage', () => {
  it('calls apiFetch on mount with query string', async () => {
    vi.mocked(apiFetch).mockResolvedValue(emptyResponse)
    wrap(<BrowsePage />)
    await waitFor(() => expect(apiFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/images?')
    ))
  })

  it('shows image count after successful fetch', async () => {
    vi.mocked(apiFetch).mockResolvedValue(oneImageResponse)
    wrap(<BrowsePage />)
    await waitFor(() => expect(screen.getByText('1 image')).toBeInTheDocument())
  })

  it('shows "No images found" when data is empty', async () => {
    vi.mocked(apiFetch).mockResolvedValue(emptyResponse)
    wrap(<BrowsePage />)
    await waitFor(() => expect(screen.getByText('No images found.')).toBeInTheDocument())
  })

  it('shows error message when fetch fails', async () => {
    vi.mocked(apiFetch).mockRejectedValue(new Error('Network error'))
    wrap(<BrowsePage />)
    await waitFor(() => expect(screen.getByText('Network error')).toBeInTheDocument())
  })

  it('Previous button is disabled at offset 0', async () => {
    vi.mocked(apiFetch).mockResolvedValue(oneImageResponse)
    wrap(<BrowsePage />)
    await waitFor(() => screen.getByText('1 image'))
    const prev = screen.getByRole('button', { name: 'Previous' })
    expect(prev).toBeDisabled()
  })
})
