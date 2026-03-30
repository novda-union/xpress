import { ChevronLeft, ShoppingCart } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface MenuHeaderProps {
  title: string
  count: number
}

export function MenuHeader({ title, count }: MenuHeaderProps) {
  const navigate = useNavigate()

  return (
    <div className="glass-panel sticky top-0 z-20 flex h-12 items-center justify-between border-b border-[var(--xp-border)] px-4">
      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => navigate(-1)}
        className="h-10 w-10 rounded-full"
      >
        <ChevronLeft className="h-6 w-6" />
      </Button>
      <p className="max-w-[14rem] truncate text-[18px] font-semibold">{title}</p>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => navigate('/cart')}
        className="relative h-10 w-10 rounded-full"
      >
        <ShoppingCart className="h-5 w-5" />
        {count > 0 ? (
          <Badge className="absolute -right-0.5 -top-0.5 flex h-5 min-w-5 items-center justify-center rounded-full px-1 text-[11px] font-semibold">
            {count}
          </Badge>
        ) : null}
      </Button>
    </div>
  )
}
