import { formatPrice } from '../../lib/format'
import { Button } from '@/components/ui/button'

interface CartBarProps {
  count: number
  total: number
  totalCartsCount: number
  onOpen: () => void
}

export function CartBar({ count, total, totalCartsCount, onOpen }: CartBarProps) {
  return (
    <div className="fixed inset-x-0 bottom-0 z-20 border-t border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-3">
      <Button
        type="button"
        onClick={onOpen}
        className="mx-auto flex h-14 w-full max-w-[32rem] items-center justify-between rounded-[20px] px-5 text-left"
      >
        <span className="flex items-center gap-2 text-sm font-semibold">
          Cart ({count})
          {totalCartsCount > 1 && (
            <span className="rounded-full bg-white/20 px-2 py-0.5 text-[11px] font-semibold">
              {totalCartsCount} carts
            </span>
          )}
        </span>
        <span className="text-base font-semibold">{formatPrice(total)} UZS</span>
      </Button>
    </div>
  )
}
