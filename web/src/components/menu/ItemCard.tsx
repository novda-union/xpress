import { Plus } from 'lucide-react'
import { formatPrice } from '../../lib/format'
import type { MenuItem } from '../../types'

interface ItemCardProps {
  item: MenuItem
  onSelect: (item: MenuItem) => void
}

export function ItemCard({ item, onSelect }: ItemCardProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect(item)}
      className="xp-card overflow-hidden text-left"
    >
      <div className="relative">
        <img
          src={item.image_url || 'https://placehold.co/320x220?text=Item'}
          alt={item.name}
          className={`h-36 w-full object-cover ${item.is_available ? '' : 'grayscale'}`}
        />
        {!item.is_available ? (
          <span className="absolute right-3 top-3 rounded-full bg-black/70 px-2 py-1 text-[11px] font-semibold text-white">
            Unavailable
          </span>
        ) : null}
      </div>
      <div className="px-3 pb-3 pt-3">
        <p className="text-[15px] font-semibold">{item.name}</p>
        <p className="mt-1 line-clamp-2 text-[13px] text-[var(--tg-theme-hint-color)]">{item.description}</p>
        <div className="mt-3 flex items-center justify-between">
          <p className="text-sm font-semibold text-[var(--xp-brand)]">{formatPrice(item.base_price)} UZS</p>
          <span className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--xp-brand)] text-white">
            <Plus className="h-4 w-4" />
          </span>
        </div>
      </div>
    </button>
  )
}
