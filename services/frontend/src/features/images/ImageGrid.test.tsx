import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'
import { ImageGrid } from './ImageGrid'
import type { ImageSummary } from './types'

const wrap = (ui: React.ReactElement) =>
  render(<MantineProvider>{ui}</MantineProvider>)

const sampleImage: ImageSummary = {
  id: '1',
  imageId: 'img-1',
  originalFilename: 'photo.webp',
  thumbnailKey: 'uploads/1/thumb.webp',
  webKey: 'uploads/1/web.webp',
  originalKey: 'uploads/1/original.webp',
  width: 1920,
  height: 1080,
  tags: ['beach', 'sunset'],
  people: ['Alice'],
  occasionCategory: null,
  uploadedAt: '2024-01-01T00:00:00Z',
}

describe('ImageGrid', () => {
  it('shows a Loader when loading', () => {
    wrap(<ImageGrid loading={true} />)
    // Mantine Loader renders an svg or div — check absence of content
    expect(screen.queryByText('No images found.')).not.toBeInTheDocument()
    expect(screen.queryByText('photo.webp')).not.toBeInTheDocument()
  })

  it('shows error message when error is provided', () => {
    wrap(<ImageGrid error="Failed to load" />)
    expect(screen.getByText('Failed to load')).toBeInTheDocument()
  })

  it('shows empty message when images array is empty', () => {
    wrap(<ImageGrid images={[]} />)
    expect(screen.getByText('No images found.')).toBeInTheDocument()
  })

  it('renders image cards when images are provided', () => {
    wrap(<ImageGrid images={[sampleImage]} />)
    expect(screen.getAllByText('photo.webp').length).toBeGreaterThan(0)
  })

  it('shows tags on image card', () => {
    wrap(<ImageGrid images={[sampleImage]} />)
    expect(screen.getByText('beach')).toBeInTheDocument()
    expect(screen.getByText('sunset')).toBeInTheDocument()
  })
})
