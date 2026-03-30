import { Clock3, Minus, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { MenuHeader } from '../components/menu/MenuHeader'
import { AppShell } from '../components/layout/AppShell'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { useCartStore } from '../store/cart'
import type { Order } from '../types'

const ETA_OPTIONS = [5, 10, 15, 20, 30]

export default function CartPage() {
  const navigate = useNavigate()
  const cart = useCartStore()
  const [eta, setEta] = useState<number>(15)
  const [loading, setLoading] = useState(false)

  async function placeOrder() {
    if (!cart.branch || cart.items.length === 0) {
      return
    }

    setLoading(true)

    try {
      const order = await api<Order>('/orders', {
        method: 'POST',
        body: JSON.stringify({
          branch_id: cart.branch.branchId,
          payment_method: 'pay_at_pickup',
          eta_minutes: eta,
          items: cart.items.map((item) => ({
            item_id: item.itemId,
            item_name: item.name,
            item_price: item.price,
            quantity: item.quantity,
            modifiers: item.modifiers.map((modifier) => ({
              modifier_id: modifier.id,
              modifier_name: modifier.name,
              price_adjustment: modifier.price,
            })),
          })),
        }),
      })
      cart.clear()
      navigate(`/order/${order.id}`)
    } catch (error) {
      alert(error instanceof Error ? error.message : 'Failed to place order')
    } finally {
      setLoading(false)
    }
  }

  if (!cart.branch || cart.items.length === 0) {
    return (
      <div className="min-h-screen px-4 pt-24 text-center">
        <p className="text-lg font-semibold">Your cart is empty</p>
        <p className="mt-2 text-sm text-[var(--tg-theme-hint-color)]">Pick something delicious from a branch nearby.</p>
        <button type="button" onClick={() => navigate('/')} className="mt-6 rounded-full bg-[var(--xp-brand)] px-5 py-3 text-sm font-semibold text-white">
          Browse branches
        </button>
      </div>
    )
  }

  return (
    <AppShell header={<MenuHeader title="Your Cart" count={cart.count()} />}>
      <div className="px-4 pb-28 pt-4">
        <div className="space-y-3">
          {cart.items.map((item, index) => (
            <div key={`${item.itemId}-${index}`} className="xp-card flex gap-3 p-4">
              <img
                src={item.imageUrl || 'https://placehold.co/96x96?text=Item'}
                alt={item.name}
                className="h-12 w-12 rounded-2xl object-cover"
              />
              <div className="min-w-0 flex-1">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="font-semibold">{item.name}</p>
                    <p className="line-clamp-2 text-sm text-[var(--tg-theme-hint-color)]">
                      {item.modifiers.map((modifier) => modifier.name).join(', ') || 'No extras'}
                    </p>
                  </div>
                  <button type="button" onClick={() => cart.removeItem(index)} className="text-[var(--tg-theme-hint-color)]">
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
                <div className="mt-3 flex items-center justify-between">
                  <div className="flex items-center gap-2 rounded-full bg-[var(--xp-card-bg)] px-2 py-1">
                    <button type="button" onClick={() => cart.updateQuantity(index, Math.max(1, item.quantity - 1))}>
                      <Minus className="h-4 w-4" />
                    </button>
                    <span className="w-6 text-center text-sm font-semibold">{item.quantity}</span>
                    <button type="button" onClick={() => cart.updateQuantity(index, item.quantity + 1)}>
                      <Plus className="h-4 w-4" />
                    </button>
                  </div>
                  <p className="font-semibold">{formatPrice(item.totalPrice)} UZS</p>
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="mt-6">
          <p className="mb-3 font-semibold">Arrive in</p>
          <div className="scrollbar-none flex gap-2 overflow-x-auto">
            {ETA_OPTIONS.map((minutes) => (
              <button
                type="button"
                key={minutes}
                onClick={() => setEta(minutes)}
                className={`xp-pill flex shrink-0 items-center gap-2 px-4 text-sm font-medium ${
                  eta === minutes ? 'bg-[var(--xp-brand)] text-white' : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
                }`}
              >
                <Clock3 className="h-4 w-4" />
                {minutes} min
              </button>
            ))}
          </div>
        </div>

        <div className="mt-6 border-t border-[var(--xp-border)] pt-4">
          <div className="flex items-center justify-between text-base font-semibold">
            <span>Subtotal</span>
            <span>{formatPrice(cart.total())} UZS</span>
          </div>
        </div>
      </div>

      <div className="fixed inset-x-0 bottom-0 z-20 border-t border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-3">
        <button
          type="button"
          onClick={placeOrder}
          disabled={loading}
          className="mx-auto flex h-14 w-full max-w-[32rem] items-center justify-center rounded-[20px] bg-[var(--xp-brand)] px-5 text-base font-semibold text-white disabled:opacity-50"
        >
          {loading ? 'Placing order...' : `Place Order — ${formatPrice(cart.total())} UZS`}
        </button>
      </div>
    </AppShell>
  )
}
