import { Routes, Route, Navigate } from 'react-router-dom'

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<div>Login placeholder</div>} />
      <Route path="/auth/verify" element={<div>Verify placeholder</div>} />
      <Route path="/" element={<div>Browse placeholder</div>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
