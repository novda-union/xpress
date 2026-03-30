import { ChevronLeft, ShoppingCart } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

interface MenuHeaderProps {
  title: string
  count: number
}

export function MenuHeader({ title, count }: MenuHeaderProps) {
  const navigate = useNavigate()

  return (
    <div className="glass-panel sticky top-0 z-20 flex h-12 items-center justify-between border-b border-[var(--xp-border)] px-4">
      <button
        type="button"
        onClick={() => navigate(-1)}
        className="flex h-11 w-11 items-center justify-center rounded-full"
      >
        <ChevronLeft className="h-6 w-6" />
      </button>
      <p className="max-w-[14rem] truncate text-[18px] font-semibold">{title}</p>
      <button
        type="button"
        onClick={() => navigate('/cart')}
        className="relative flex h-11 w-11 items-center justify-center rounded-full"
      >
        <ShoppingCart className="h-5 w-5" />
        {count > 0 ? (
          <span className="absolute right-1 top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-[var(--xp-brand)] px-1 text-[11px] font-semibold text-white">
            {count}
          </span>
        ) : null}
      </button>
    </div>
  )
}
