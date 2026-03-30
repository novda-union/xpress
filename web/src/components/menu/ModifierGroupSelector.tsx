import type { ModifierGroup } from '../../types'

interface ModifierGroupSelectorProps {
  group: ModifierGroup
  selected: string[]
  onToggle: (group: ModifierGroup, modifierId: string) => void
}

export function ModifierGroupSelector({ group, selected, onToggle }: ModifierGroupSelectorProps) {
  return (
    <section className="mt-6">
      <div className="mb-3 border-b border-[var(--xp-border)] pb-2">
        <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-[var(--tg-theme-hint-color)]">
          {group.name}
          {group.is_required ? ' · Required' : ''}
        </p>
      </div>
      <div className="space-y-2">
        {group.modifiers.map((modifier) => {
          const active = selected.includes(modifier.id)
          return (
            <button
              type="button"
              key={modifier.id}
              onClick={() => onToggle(group, modifier.id)}
              className={`flex min-h-[52px] w-full items-center justify-between rounded-2xl border px-4 py-3.5 text-left ${
                active
                  ? 'border-[var(--xp-brand)] bg-[var(--xp-brand-muted)]'
                  : 'border-[var(--xp-border)] bg-[var(--xp-card-bg)]'
              }`}
            >
              <div className="flex items-center gap-3">
                <span
                  className={`flex h-5 w-5 items-center justify-center rounded-full border ${
                    active ? 'border-[var(--xp-brand)] bg-[var(--xp-brand)]' : 'border-[var(--xp-border)]'
                  }`}
                >
                  <span className={`h-2 w-2 rounded-full ${active ? 'bg-white' : 'bg-transparent'}`} />
                </span>
                <span className="text-sm">{modifier.name}</span>
              </div>
              <span className="text-sm text-[var(--tg-theme-hint-color)]">+{modifier.price_adjustment.toLocaleString('en-US')} UZS</span>
            </button>
          )
        })}
      </div>
    </section>
  )
}
