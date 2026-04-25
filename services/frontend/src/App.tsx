import { Routes, Route, Navigate } from 'react-router-dom'
import { LoginPage } from '@/features/auth/LoginPage'
import { VerifyPage } from '@/features/auth/VerifyPage'

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/auth/verify" element={<VerifyPage />} />
      <Route path="/" element={<div>Browse placeholder</div>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
