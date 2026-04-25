import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Center, Loader, Alert, Stack, Text } from '@mantine/core'
import { useAuthStore } from './authStore'
import { apiFetch, ApiError } from '@/api/client'

export function VerifyPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const fetchCurrentUser = useAuthStore((s) => s.fetchCurrentUser)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const token = params.get('token')
    if (!token) {
      setError('Missing token in URL.')
      return
    }
    apiFetch(`/api/v1/auth/verify?token=${encodeURIComponent(token)}`)
      .then(async () => {
        await fetchCurrentUser()
        navigate('/', { replace: true })
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError) {
          setError(err.detail)
        } else {
          setError('Verification failed. The link may have expired.')
        }
      })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  if (error) {
    return (
      <Center h="100vh">
        <Stack align="center">
          <Alert color="red" title="Sign-in failed">{error}</Alert>
          <Text
            component="a"
            href="/login"
            c="blue"
            size="sm"
          >
            Try again
          </Text>
        </Stack>
      </Center>
    )
  }

  return (
    <Center h="100vh">
      <Loader />
    </Center>
  )
}
