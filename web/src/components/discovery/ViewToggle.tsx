import { LayoutList, Map } from 'lucide-react'

interface ViewToggleProps {
  value: 'map' | 'list'
  onChange: (value: 'map' | 'list') => void
}

export function ViewToggle({ value, onChange }: ViewToggleProps) {
  return (
    <div
      className="fixed bottom-5 left-1/2 z-20 -translate-x-1/2"
      style={{
        background: 'color-mix(in srgb, var(--xp-card-bg) 72%, transparent)',
        backdropFilter: 'blur(28px) saturate(200%)',
        WebkitBackdropFilter: 'blur(28px) saturate(200%)',
        border: '1px solid color-mix(in srgb, var(--xp-border) 60%, transparent)',
        borderRadius: '9999px',
        boxShadow: '0 4px 20px rgba(0,0,0,0.14), 0 1px 4px rgba(0,0,0,0.08)',
        padding: '4px',
      }}
    >
      <div className="relative flex">
        {/* Sliding active indicator */}
        <span
          className="pointer-events-none absolute inset-y-0 w-1/2 rounded-full"
          style={{
            background: 'var(--xp-brand)',
            transform: value === 'map' ? 'translateX(0%)' : 'translateX(100%)',
            transition: 'transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1)',
          }}
        />
        <button
          type="button"
          onClick={() => onChange('map')}
          className="relative z-10 flex h-9 w-12 items-center justify-center rounded-full"
          style={{
            color: value === 'map' ? '#fff' : 'var(--tg-theme-hint-color)',
            transition: 'color 0.25s ease',
          }}
          aria-label="Map view"
        >
          <Map className="h-4 w-4" />
        </button>
        <button
          type="button"
          onClick={() => onChange('list')}
          className="relative z-10 flex h-9 w-12 items-center justify-center rounded-full"
          style={{
            color: value === 'list' ? '#fff' : 'var(--tg-theme-hint-color)',
            transition: 'color 0.25s ease',
          }}
          aria-label="List view"
        >
          <LayoutList className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}
