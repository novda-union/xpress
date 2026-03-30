import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../components/common/StatusBadge'
import { MenuHeader } from '../components/menu/MenuHeader'
import { AppShell } from '../components/layout/AppShell'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import type { Order } from '../types'

export default function OrdersPage() {
  const navigate = useNavigate()
  const [orders, setOrders] = useState<Order[]>([])

  useEffect(() => {
    api<Order[]>('/orders')
      .then(setOrders)
      .catch(() => setOrders([]))
  }, [])

  return (
    <AppShell header={<MenuHeader title="My Orders" count={0} />}>
      <div className="px-4 pb-10 pt-4">
        {orders.length === 0 ? (
          <div className="xp-card px-6 py-10 text-center">
            <p className="text-lg font-semibold">No orders yet</p>
            <p className="mt-2 text-sm text-[var(--tg-theme-hint-color)]">
              Once you place an order, it will appear here.
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {orders.map((order) => (
              <button
                type="button"
                key={order.id}
                onClick={() => navigate(`/order/${order.id}`)}
                className="xp-card w-full px-4 py-4 text-left"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="font-semibold">Order #{order.order_number}</p>
                    <p className="mt-1 text-sm text-[var(--tg-theme-hint-color)]">
                      {new Date(order.created_at).toLocaleDateString()} · {formatPrice(order.total_price)} UZS
                    </p>
                  </div>
                  <StatusBadge status={order.status} />
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </AppShell>
  )
}
