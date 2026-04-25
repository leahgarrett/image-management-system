import { describe, it, expect, beforeEach } from 'vitest'
import { useFilterStore } from './filterStore'

beforeEach(() => {
  useFilterStore.setState({ tags: [], people: '', occasion: '', offset: 0 })
})

describe('useFilterStore', () => {
  it('builds empty query string when no filters set', () => {
    const qs = useFilterStore.getState().toQueryString()
    expect(qs).toBe('limit=20&offset=0')
  })

  it('includes tags when set', () => {
    useFilterStore.setState({ tags: ['beach', 'sunset'] })
    const qs = useFilterStore.getState().toQueryString()
    expect(qs).toContain('tags=beach%2Csunset')
  })

  it('includes people when set', () => {
    useFilterStore.setState({ people: 'Alice' })
    const qs = useFilterStore.getState().toQueryString()
    expect(qs).toContain('people=Alice')
  })

  it('includes occasion when set', () => {
    useFilterStore.setState({ occasion: 'birthday' })
    const qs = useFilterStore.getState().toQueryString()
    expect(qs).toContain('occasion=birthday')
  })

  it('resetFilters clears everything and resets offset', () => {
    useFilterStore.setState({ tags: ['x'], people: 'Bob', occasion: 'wedding', offset: 40 })
    useFilterStore.getState().resetFilters()
    const state = useFilterStore.getState()
    expect(state.tags).toEqual([])
    expect(state.people).toBe('')
    expect(state.occasion).toBe('')
    expect(state.offset).toBe(0)
  })

  it('setTags resets offset to 0', () => {
    useFilterStore.setState({ offset: 40 })
    useFilterStore.getState().setTags(['new'])
    expect(useFilterStore.getState().offset).toBe(0)
  })

  it('nextPage increments offset by PAGE_SIZE (20)', () => {
    useFilterStore.setState({ offset: 0 })
    useFilterStore.getState().nextPage()
    expect(useFilterStore.getState().offset).toBe(20)
  })

  it('prevPage decrements offset but not below 0', () => {
    useFilterStore.setState({ offset: 20 })
    useFilterStore.getState().prevPage()
    expect(useFilterStore.getState().offset).toBe(0)
    useFilterStore.getState().prevPage()
    expect(useFilterStore.getState().offset).toBe(0)
  })
})
