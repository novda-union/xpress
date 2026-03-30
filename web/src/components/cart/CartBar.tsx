import { formatPrice } from '../../lib/format'
import { Button } from '@/components/ui/button'

interface CartBarProps {
  count: number
  total: number
  onOpen: () => void
}

export function CartBar({ count, total, onOpen }: CartBarProps) {
  return (
    <div className="fixed inset-x-0 bottom-0 z-20 border-t border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-3">
      <Button
        type="button"
        onClick={onOpen}
        className="mx-auto flex h-14 w-full max-w-[32rem] items-center justify-between rounded-[20px] px-5 text-left"
      >
        <span className="text-sm font-semibold">Cart ({count})</span>
        <span className="text-base font-semibold">{formatPrice(total)} UZS</span>
      </Button>
    </div>
  )
}
