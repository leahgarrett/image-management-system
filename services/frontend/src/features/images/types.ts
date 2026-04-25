export interface ImageSummary {
  id: string
  imageId: string
  originalFilename: string
  thumbnailKey: string
  webKey: string
  originalKey: string
  width: number
  height: number
  tags: string[]
  people: string[]
  occasionCategory: string | null
  uploadedAt: string
}

export interface ImageListResponse {
  data: ImageSummary[]
  pagination: {
    total: number
    limit: number
    offset: number
    hasMore: boolean
  }
}
