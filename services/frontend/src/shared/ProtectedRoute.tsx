import { Navigate, Outlet } from 'react-router-dom'
import { Center, Loader } from '@mantine/core'
import { useAuthStore } from '@/features/auth/authStore'

export function ProtectedRoute() {
  const user = useAuthStore((s) => s.user)
  const isInitialising = useAuthStore((s) => s.isInitialising)
  if (isInitialising) return <Center h="100vh"><Loader /></Center>
  if (!user) return <Navigate to="/login" replace />
  return <Outlet />
}
