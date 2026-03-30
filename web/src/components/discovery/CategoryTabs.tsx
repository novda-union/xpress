import type { DiscoveryCategory } from '../../hooks/useDiscovery'

const CATEGORY_OPTIONS: Array<{ value: DiscoveryCategory; label: string }> = [
  { value: 'all', label: 'All' },
  { value: 'bar', label: 'Bars' },
  { value: 'cafe', label: 'Cafes' },
  { value: 'coffee', label: 'Coffee' },
  { value: 'restaurant', label: 'Restaurants' },
  { value: 'fastfood', label: 'Fast Food' },
]

interface CategoryTabsProps {
  value: DiscoveryCategory
  onChange: (value: DiscoveryCategory) => void
}

export function CategoryTabs({ value, onChange }: CategoryTabsProps) {
  return (
    <div className="scrollbar-none flex gap-2 overflow-x-auto px-4 py-3">
      {CATEGORY_OPTIONS.map((option) => (
        <button
          key={option.value}
          type="button"
          onClick={() => onChange(option.value)}
          className={`xp-pill whitespace-nowrap px-4 text-[13px] font-medium ${
            value === option.value
              ? 'bg-[var(--xp-brand)] text-white'
              : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
          }`}
        >
          {option.label}
        </button>
      ))}
    </div>
  )
}
