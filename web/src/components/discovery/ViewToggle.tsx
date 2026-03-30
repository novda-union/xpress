import { LayoutList, Map } from 'lucide-react'

interface ViewToggleProps {
  value: 'map' | 'list'
  onChange: (value: 'map' | 'list') => void
}

export function ViewToggle({ value, onChange }: ViewToggleProps) {
  return (
    <div className="glass-panel flex rounded-full p-1 shadow-sm">
      <button
        type="button"
        onClick={() => onChange('map')}
        className={`xp-pill flex items-center gap-1.5 px-4 py-1.5 text-[13px] font-medium ${
          value === 'map' ? 'bg-[var(--xp-brand)] text-white' : 'text-[var(--tg-theme-hint-color)]'
        }`}
      >
        <Map className="h-4 w-4" />
        Map
      </button>
      <button
        type="button"
        onClick={() => onChange('list')}
        className={`xp-pill flex items-center gap-1.5 px-4 py-1.5 text-[13px] font-medium ${
          value === 'list' ? 'bg-[var(--xp-brand)] text-white' : 'text-[var(--tg-theme-hint-color)]'
        }`}
      >
        <LayoutList className="h-4 w-4" />
        List
      </button>
    </div>
  )
}
