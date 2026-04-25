import { AppShell as MantineAppShell, Group, Text, Button } from '@mantine/core'
import { Outlet } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/authStore'

export function AppShell() {
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)

  return (
    <MantineAppShell
      header={{ height: 56 }}
      padding="md"
    >
      <MantineAppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Text fw={700} size="lg">Family Photos</Text>
          <Group>
            <Text size="sm" c="dimmed">{user?.email}</Text>
            <Button variant="subtle" size="sm" onClick={logout}>Sign out</Button>
          </Group>
        </Group>
      </MantineAppShell.Header>
      <MantineAppShell.Main>
        <Outlet />
      </MantineAppShell.Main>
    </MantineAppShell>
  )
}
