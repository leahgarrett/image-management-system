import { SimpleGrid, Text, Center, Loader } from '@mantine/core'
import { ImageCard } from './ImageCard'
import type { ImageSummary } from './types'

type Props =
  | { loading: true; error?: never; images?: never }
  | { loading?: false; error: string; images?: never }
  | { loading?: false; error?: null; images: ImageSummary[] }

export function ImageGrid({ images, loading, error }: Props) {
  if (loading) {
    return (
      <Center h={200}>
        <Loader />
      </Center>
    )
  }
  if (error) {
    return (
      <Center h={200}>
        <Text c="red">{error}</Text>
      </Center>
    )
  }
  if (!images || images.length === 0) {
    return (
      <Center h={200}>
        <Text c="dimmed">No images found.</Text>
      </Center>
    )
  }
  return (
    <SimpleGrid cols={{ base: 1, sm: 2, md: 3, lg: 4 }} spacing="md">
      {images.map((img) => (
        <ImageCard key={img.id} image={img} />
      ))}
    </SimpleGrid>
  )
}
