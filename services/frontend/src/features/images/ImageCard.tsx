import { Card, Text, Badge, Group, Stack } from '@mantine/core'
import type { ImageSummary } from './types'

interface Props {
  image: ImageSummary
}

export function ImageCard({ image }: Props) {
  return (
    <Card shadow="sm" padding="sm" radius="md" withBorder>
      <Card.Section>
        <Stack
          h={180}
          align="center"
          justify="center"
          bg="gray.1"
          style={{ overflow: 'hidden' }}
        >
          <Text size="xs" c="dimmed" ta="center" px="xs">
            {image.originalFilename}
          </Text>
        </Stack>
      </Card.Section>
      <Stack mt="sm" gap="xs">
        <Text size="sm" fw={500} lineClamp={1}>
          {image.originalFilename}
        </Text>
        {image.tags.length > 0 && (
          <Group gap="xs">
            {image.tags.slice(0, 3).map((tag) => (
              <Badge key={tag} size="xs" variant="light">
                {tag}
              </Badge>
            ))}
            {image.tags.length > 3 && (
              <Badge size="xs" variant="outline">
                +{image.tags.length - 3}
              </Badge>
            )}
          </Group>
        )}
        {image.people.length > 0 && (
          <Text size="xs" c="dimmed">
            {image.people.join(', ')}
          </Text>
        )}
      </Stack>
    </Card>
  )
}
