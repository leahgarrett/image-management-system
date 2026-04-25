import { create } from 'zustand'
import { apiFetch, ApiError } from '@/api/client'

export interface User {
  id: string
  email: string
  role: 'admin' | 'contributor'
}

interface AuthState {
  user: User | null
  isInitialising: boolean
  requestMagicLink: (email: string) => Promise<void>
  fetchCurrentUser: () => Promise<void>
  logout: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isInitialising: true,

  requestMagicLink: async (email: string) => {
    await apiFetch('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email }),
    })
  },

  fetchCurrentUser: async () => {
    try {
      const user = await apiFetch<User>('/api/v1/users/me')
      if (user !== undefined) set({ user, isInitialising: false })
      else set({ isInitialising: false })
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        set({ user: null, isInitialising: false })
      } else {
        set({ isInitialising: false })
      }
    }
  },

  logout: async () => {
    set({ user: null })
    try {
      await apiFetch('/api/v1/auth/logout', { method: 'POST' })
    } catch {
      // cookie may remain valid server-side; local state cleared regardless
    }
  },
}))
