import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'
import { useFilterStore } from './filterStore'
import { FilterBar } from './FilterBar'

const wrap = (ui: React.ReactElement) =>
  render(<MantineProvider>{ui}</MantineProvider>)

beforeEach(() => {
  useFilterStore.setState({ tags: [], people: '', occasion: '', offset: 0 })
})

describe('FilterBar', () => {
  it('renders without crashing', () => {
    wrap(<FilterBar />)
    // Mantine TagsInput renders multiple elements with the Tags label; use getAllByLabelText
    expect(screen.getAllByLabelText('Tags').length).toBeGreaterThan(0)
    expect(screen.getByLabelText('People')).toBeInTheDocument()
    expect(screen.getAllByLabelText('Occasion').length).toBeGreaterThan(0)
  })

  it('does not show Clear filters when no filters are set', () => {
    wrap(<FilterBar />)
    expect(screen.queryByText('Clear filters')).not.toBeInTheDocument()
  })

  it('shows Clear filters button when people filter is set', () => {
    useFilterStore.setState({ people: 'Alice' })
    wrap(<FilterBar />)
    expect(screen.getByText('Clear filters')).toBeInTheDocument()
  })

  it('calls resetFilters when Clear filters is clicked', () => {
    useFilterStore.setState({ people: 'Alice' })
    wrap(<FilterBar />)
    fireEvent.click(screen.getByText('Clear filters'))
    expect(useFilterStore.getState().people).toBe('')
    expect(useFilterStore.getState().tags).toEqual([])
    expect(useFilterStore.getState().occasion).toBe('')
  })

  it('updates people store on TextInput change', () => {
    wrap(<FilterBar />)
    fireEvent.change(screen.getByLabelText('People'), { target: { value: 'Bob' } })
    expect(useFilterStore.getState().people).toBe('Bob')
  })
})
