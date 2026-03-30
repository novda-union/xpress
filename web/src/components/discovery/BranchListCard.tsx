import { MapPin } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { formatDistance } from '../../lib/distance'
import type { DiscoverBranch } from '../../types'

interface BranchListCardProps {
  branch: DiscoverBranch & { distanceKm: number }
  onSelect: (branch: DiscoverBranch & { distanceKm: number }) => void
}

const badgeClasses: Record<string, string> = {
  bar: 'bg-purple-500/15 text-purple-600 hover:bg-purple-500/15',
  cafe: 'bg-amber-500/15 text-amber-600 hover:bg-amber-500/15',
  coffee: 'bg-orange-900/10 text-orange-700 hover:bg-orange-900/10',
  restaurant: 'bg-green-500/15 text-green-600 hover:bg-green-500/15',
  fastfood: 'bg-red-500/15 text-red-600 hover:bg-red-500/15',
}

export function BranchListCard({ branch, onSelect }: BranchListCardProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect(branch)}
      className="xp-card flex w-full cursor-pointer gap-4 p-4 text-left transition-transform hover:-translate-y-0.5 hover:opacity-95"
    >
      <img
        src={branch.store_logo_url || 'https://placehold.co/120x120?text=XG'}
        alt={branch.store_name}
        className="h-[72px] w-[72px] rounded-2xl object-cover"
      />
      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-[15px] font-semibold">{branch.store_name}</p>
            <p className="mt-1 text-[13px] text-[var(--tg-theme-hint-color)]">{branch.branch_name}</p>
          </div>
          <span className="text-[13px] font-medium text-[var(--tg-theme-hint-color)]">
            {formatDistance(branch.distanceKm)}
          </span>
        </div>
        <p className="mt-2 flex items-center gap-1.5 text-[13px] text-[var(--tg-theme-hint-color)]">
          <MapPin className="h-3.5 w-3.5" />
          <span className="line-clamp-2">{branch.branch_address}</span>
        </p>
        <Badge
          variant="secondary"
          className={`mt-3 text-[11px] ${badgeClasses[branch.store_category] ?? ''}`}
        >
          {branch.store_category}
        </Badge>
      </div>
    </button>
  )
}
