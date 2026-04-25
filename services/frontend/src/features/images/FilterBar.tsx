import { Group, TagsInput, TextInput, Select, Button } from '@mantine/core'
import { useFilterStore } from './filterStore'

const OCCASION_OPTIONS = [
  { value: '', label: 'Any occasion' },
  { value: 'birthday', label: 'Birthday' },
  { value: 'wedding', label: 'Wedding' },
  { value: 'graduation', label: 'Graduation' },
  { value: 'holiday', label: 'Holiday' },
  { value: 'vacation', label: 'Vacation' },
  { value: 'work_event', label: 'Work event' },
  { value: 'party', label: 'Party' },
  { value: 'family_gathering', label: 'Family gathering' },
  { value: 'sports_event', label: 'Sports event' },
  { value: 'concert', label: 'Concert' },
  { value: 'conference', label: 'Conference' },
  { value: 'ceremony', label: 'Ceremony' },
  { value: 'casual', label: 'Casual' },
  { value: 'other', label: 'Other' },
]

export function FilterBar() {
  const tags = useFilterStore((s) => s.tags)
  const people = useFilterStore((s) => s.people)
  const occasion = useFilterStore((s) => s.occasion)
  const setTags = useFilterStore((s) => s.setTags)
  const setPeople = useFilterStore((s) => s.setPeople)
  const setOccasion = useFilterStore((s) => s.setOccasion)
  const resetFilters = useFilterStore((s) => s.resetFilters)

  const hasFilters = tags.length > 0 || people !== '' || occasion !== ''

  return (
    <Group align="flex-end" mb="md" wrap="wrap">
      <TagsInput
        label="Tags"
        placeholder="Add tag"
        value={tags}
        onChange={setTags}
        style={{ minWidth: 200 }}
      />
      <TextInput
        label="People"
        placeholder="Person name"
        value={people}
        onChange={(e) => setPeople(e.currentTarget.value)}
        style={{ minWidth: 160 }}
      />
      <Select
        label="Occasion"
        data={OCCASION_OPTIONS}
        value={occasion}
        onChange={(v) => setOccasion(v ?? '')}
        style={{ minWidth: 180 }}
      />
      {hasFilters && (
        <Button variant="subtle" color="gray" onClick={resetFilters}>
          Clear filters
        </Button>
      )}
    </Group>
  )
}
