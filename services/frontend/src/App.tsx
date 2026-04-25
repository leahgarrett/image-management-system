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
