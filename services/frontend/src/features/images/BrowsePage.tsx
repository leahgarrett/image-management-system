import { useEffect, useState } from 'react'
import { Stack, Group, Button, Text } from '@mantine/core'
import { apiFetch } from '@/api/client'
import { useFilterStore } from './filterStore'
import { FilterBar } from './FilterBar'
import { ImageGrid } from './ImageGrid'
import type { ImageListResponse } from './types'

export function BrowsePage() {
  const qs = useFilterStore((s) => s.toQueryString())
  const offset = useFilterStore((s) => s.offset)
  const nextPage = useFilterStore((s) => s.nextPage)
  const prevPage = useFilterStore((s) => s.prevPage)

  const [response, setResponse] = useState<ImageListResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    apiFetch<ImageListResponse>(`/api/v1/images?${qs}`)
      .then((data) => {
        if (!cancelled) {
          setResponse(data ?? null)
          setLoading(false)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load images')
          setLoading(false)
        }
      })
    return () => { cancelled = true }
  }, [qs])

  const images = response?.data ?? []
  const pagination = response?.pagination

  return (
    <Stack>
      <FilterBar />
      {loading ? (
        <ImageGrid loading={true} />
      ) : error ? (
        <ImageGrid error={error} />
      ) : (
        <ImageGrid images={images} />
      )}
      {pagination && (
        <Group justify="space-between">
          <Text size="sm" c="dimmed">
            {pagination.total} image{pagination.total !== 1 ? 's' : ''}
          </Text>
          <Group>
            <Button
              variant="subtle"
              size="sm"
              disabled={offset === 0}
              onClick={prevPage}
            >
              Previous
            </Button>
            <Button
              variant="subtle"
              size="sm"
              disabled={!pagination.hasMore}
              onClick={nextPage}
            >
              Next
            </Button>
          </Group>
        </Group>
      )}
    </Stack>
  )
}
