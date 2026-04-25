import { useState } from 'react'
import {
  Center,
  Stack,
  Title,
  Text,
  TextInput,
  Button,
  Alert,
  Paper,
} from '@mantine/core'
import { useAuthStore } from './authStore'

export function LoginPage() {
  const requestMagicLink = useAuthStore((s) => s.requestMagicLink)
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      await requestMagicLink(email)
      setSent(true)
    } catch {
      setError('Could not send magic link. Check the email address and try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Center h="100vh">
      <Paper shadow="md" p="xl" w={400}>
        <Stack>
          <Title order={2}>Family Photos</Title>
          {sent ? (
            <Alert color="green" title="Check your email">
              We sent a sign-in link to <strong>{email}</strong>. Click the link to continue.
            </Alert>
          ) : (
            <form onSubmit={handleSubmit}>
              <Stack>
                <Text c="dimmed" size="sm">
                  Enter your email to receive a sign-in link.
                </Text>
                {error && <Alert color="red">{error}</Alert>}
                <TextInput
                  label="Email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.currentTarget.value)}
                  required
                  autoFocus
                />
                <Button type="submit" loading={loading} fullWidth>
                  Send link
                </Button>
              </Stack>
            </form>
          )}
        </Stack>
      </Paper>
    </Center>
  )
}
