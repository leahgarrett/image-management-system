import { create } from 'zustand'

const PAGE_SIZE = 20

interface FilterState {
  tags: string[]
  people: string
  occasion: string
  offset: number
  setTags: (tags: string[]) => void
  setPeople: (people: string) => void
  setOccasion: (occasion: string) => void
  nextPage: () => void
  prevPage: () => void
  resetFilters: () => void
  toQueryString: () => string
}

export const useFilterStore = create<FilterState>((set, get) => ({
  tags: [],
  people: '',
  occasion: '',
  offset: 0,

  setTags: (tags) => set({ tags, offset: 0 }),
  setPeople: (people) => set({ people, offset: 0 }),
  setOccasion: (occasion) => set({ occasion, offset: 0 }),
  nextPage: () => set((s) => ({ offset: s.offset + PAGE_SIZE })),
  prevPage: () => set((s) => ({ offset: Math.max(0, s.offset - PAGE_SIZE) })),
  resetFilters: () => set({ tags: [], people: '', occasion: '', offset: 0 }),

  toQueryString: () => {
    const { tags, people, occasion, offset } = get()
    const params = new URLSearchParams()
    if (tags.length > 0) params.set('tags', tags.join(','))
    if (people) params.set('people', people)
    if (occasion) params.set('occasion', occasion)
    params.set('limit', String(PAGE_SIZE))
    params.set('offset', String(offset))
    return params.toString()
  },
}))
