import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuthStore } from './authStore'

vi.mock('@/api/client', () => ({
  apiFetch: vi.fn(),
  ApiError: class ApiError extends Error {
    constructor(
      public status: number,
      public detail: string,
      public type: string = 'error',
    ) { super(detail) }
  },
}))

import { apiFetch } from '@/api/client'

beforeEach(() => {
  useAuthStore.setState({ user: null, isInitialising: false })
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

  it('sets isInitialising to false after fetchCurrentUser resolves', async () => {
    useAuthStore.setState({ isInitialising: true })
    vi.mocked(apiFetch).mockResolvedValue({ id: 'u1', email: 'a@b.com', role: 'admin' })
    await useAuthStore.getState().fetchCurrentUser()
    expect(useAuthStore.getState().isInitialising).toBe(false)
  })

  it('sets isInitialising to false on 401', async () => {
    useAuthStore.setState({ isInitialising: true })
    const { ApiError } = await import('@/api/client')
    vi.mocked(apiFetch).mockRejectedValue(new ApiError(401, 'Unauthorized'))
    await useAuthStore.getState().fetchCurrentUser()
    expect(useAuthStore.getState().isInitialising).toBe(false)
  })
})

describe('requestMagicLink', () => {
  it('calls POST /auth/login with email', async () => {
    vi.mocked(apiFetch).mockResolvedValue(undefined)
    await useAuthStore.getState().requestMagicLink('user@example.com')
    expect(apiFetch).toHaveBeenCalledWith('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email: 'user@example.com' }),
    })
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
