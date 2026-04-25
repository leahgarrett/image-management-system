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

  it('returns undefined for 204 No Content', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValue(new Response(null, { status: 204 }))
    const result = await apiFetch('/api/v1/images/x')
    expect(result).toBeUndefined()
  })
})
