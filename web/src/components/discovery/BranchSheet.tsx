import { ArrowRight, MapPin, X } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { formatPrice } from '../../lib/format'
import type { DiscoverBranch } from '../../types'

interface BranchSheetProps {
  branch: (DiscoverBranch & { distanceKm: number }) | null
  onClose: () => void
}

export function BranchSheet({ branch, onClose }: BranchSheetProps) {
  const navigate = useNavigate()

  if (!branch) {
    return null
  }

  return (
    <div className="fixed inset-0 z-30">
      <button
        type="button"
        aria-label="Close sheet"
        onClick={onClose}
        className="absolute inset-0 bg-[var(--xp-overlay)]"
      />
      <div className="absolute inset-x-0 bottom-0 z-10 rounded-t-[20px] bg-[var(--tg-theme-bg-color)] px-4 pb-safe shadow-[var(--shadow-lg)]">
        <div className="mx-auto mt-3 h-1 w-8 rounded-full bg-[var(--tg-theme-hint-color)]/40" />
        <img
          src={branch.banner_image_url || 'https://placehold.co/800x320?text=Xpressgo'}
          alt={branch.branch_name}
          className="mt-4 h-40 w-full rounded-[20px] object-cover"
        />
        <div className="mt-4 flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold">{branch.store_name}</h2>
            <p className="text-sm text-[var(--tg-theme-hint-color)]">{branch.branch_name}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--xp-card-bg)]"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <p className="mt-2 flex items-center gap-1.5 text-sm text-[var(--tg-theme-hint-color)]">
          <MapPin className="h-4 w-4" />
          {branch.branch_address}
        </p>

        <div className="scrollbar-none mt-4 flex gap-3 overflow-x-auto pb-1">
          {branch.preview_items.map((item) => (
            <button
              type="button"
              key={item.id}
              onClick={() => navigate(`/item/${item.id}?branch=${branch.branch_id}`)}
              className="xp-card w-[100px] shrink-0 overflow-hidden text-left"
            >
              <img
                src={item.image_url || 'https://placehold.co/160x120?text=Item'}
                alt={item.name}
                className="h-20 w-full object-cover"
              />
              <div className="px-2 pb-2 pt-2">
                <p className="line-clamp-2 text-xs font-medium">{item.name}</p>
                <p className="mt-1 text-xs font-semibold text-[var(--xp-brand)]">{formatPrice(item.base_price)} UZS</p>
              </div>
            </button>
          ))}
        </div>

        <button
          type="button"
          onClick={() => navigate(`/branch/${branch.branch_id}`)}
          className="mt-5 flex h-[52px] w-full items-center justify-center gap-2 rounded-[20px] bg-[var(--xp-brand)] px-4 text-sm font-semibold text-white"
        >
          See Full Menu
          <ArrowRight className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}
