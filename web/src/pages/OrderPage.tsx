import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { StatusBadge } from '../components/common/StatusBadge'
import { MenuHeader } from '../components/menu/MenuHeader'
import { AppShell } from '../components/layout/AppShell'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { useWebSocket } from '../hooks/useWebSocket'
import type { BranchDetail, Order } from '../types'

const STATUS_STEPS = ['pending', 'accepted', 'preparing', 'ready', 'picked_up']

export default function OrderPage() {
  const { id } = useParams<{ id: string }>()
  const [order, setOrder] = useState<Order | null>(null)
  const [detail, setDetail] = useState<BranchDetail | null>(null)

  useEffect(() => {
    if (!id) {
      return
    }
    api<Order>(`/orders/${id}`).then((nextOrder) => {
      setOrder(nextOrder)
      api<BranchDetail>(`/branches/${nextOrder.branch_id}`).then(setDetail).catch(() => undefined)
    })
  }, [id])

  useWebSocket((message) => {
    if (message.order_id === id && id) {
      api<Order>(`/orders/${id}`).then(setOrder)
    }
  })

  async function cancelOrder() {
    if (!order) {
      return
    }
    const updated = await api<Order>(`/orders/${order.id}/cancel`, { method: 'PUT' })
    setOrder(updated)
  }

  if (!order) {
    return (
      <AppShell header={<MenuHeader title="Order" count={0} />}>
        <div className="animate-pulse px-4 pt-6 space-y-4">
          <div className="xp-card h-40 p-5" />
          <div className="xp-card h-32 p-5" />
          <div className="xp-card h-20 p-5" />
        </div>
      </AppShell>
    )
  }

  const activeStep = STATUS_STEPS.indexOf(order.status)
  const canCancel = order.status === 'pending'

  return (
    <AppShell header={<MenuHeader title={`Order #${order.order_number}`} count={0} />}>
      <div className="px-4 pb-10 pt-6">
        <div className="xp-card p-5 text-center">
          <StatusBadge status={order.status} />
          <p className="mt-4 text-lg font-semibold">Your order is on its way</p>
          <p className="mt-2 text-sm text-[var(--tg-theme-hint-color)]">
            {detail ? `${detail.store.name} · ${detail.branch.name}` : 'Tracking your branch order'}
          </p>

          <div className="mt-6 flex items-start justify-between gap-2">
            {STATUS_STEPS.map((step, index) => (
              <div key={step} className="flex flex-1 flex-col items-center gap-2">
                <div className={`h-3 w-3 rounded-full ${index <= activeStep ? 'bg-[var(--xp-brand)]' : 'bg-[var(--xp-border)]'}`} />
                <span className="text-[10px] text-[var(--tg-theme-hint-color)]">{step.replace('_', ' ')}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="xp-card mt-4 p-5">
          <p className="font-semibold">Items</p>
          <div className="mt-4 space-y-3">
            {order.items.map((item) => (
              <div key={item.id} className="flex justify-between gap-3 text-sm">
                <div>
                  <p>{item.quantity}× {item.item_name}</p>
                  {item.modifiers.length > 0 ? (
                    <p className="text-xs text-[var(--tg-theme-hint-color)]">
                      {item.modifiers.map((modifier) => modifier.modifier_name).join(', ')}
                    </p>
                  ) : null}
                </div>
                <span>{formatPrice(item.item_price * item.quantity)} UZS</span>
              </div>
            ))}
          </div>
          <div className="mt-4 border-t border-[var(--xp-border)] pt-4 flex justify-between font-semibold">
            <span>Total</span>
            <span>{formatPrice(order.total_price)} UZS</span>
          </div>
        </div>

        <div className="xp-card mt-4 p-5">
          <p className="text-sm text-[var(--tg-theme-hint-color)]">Estimated arrival: ~{order.eta_minutes} min</p>
          <p className="mt-2 text-sm text-[var(--tg-theme-hint-color)]">Payment: Pay at pickup</p>
        </div>

        {canCancel ? (
          <button
            type="button"
            onClick={() => void cancelOrder()}
            className="mt-4 flex h-12 w-full items-center justify-center rounded-[20px] border border-[var(--xp-border)] bg-transparent text-sm font-semibold"
          >
            Cancel Order
          </button>
        ) : null}
      </div>
    </AppShell>
  )
}
