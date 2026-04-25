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
