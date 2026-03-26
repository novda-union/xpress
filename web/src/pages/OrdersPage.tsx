import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import type { Order } from '../types'

const STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  accepted: 'Accepted',
  preparing: 'Preparing',
  ready: 'Ready',
  picked_up: 'Picked Up',
  rejected: 'Rejected',
  cancelled: 'Cancelled',
}

export default function OrdersPage() {
  const navigate = useNavigate()
  const [orders, setOrders] = useState<Order[]>([])

  useEffect(() => {
    api<Order[]>('/orders').catch(() => []).then(setOrders)
  }, [])

  return (
    <div className="max-w-lg mx-auto p-4">
      <h1 className="text-xl font-bold mb-4">My Orders</h1>

      {orders.length === 0 ? (
        <p className="text-gray-500 text-center mt-10">No orders yet</p>
      ) : (
        <div className="space-y-3">
          {orders.map((order) => (
            <button
              key={order.id}
              onClick={() => navigate(`/order/${order.id}`)}
              className="w-full text-left bg-white rounded-lg p-4 shadow-sm border"
            >
              <div className="flex justify-between items-start">
                <span className="font-bold">#{order.order_number}</span>
                <span className="text-xs px-2 py-1 bg-gray-100 rounded">
                  {STATUS_LABELS[order.status]}
                </span>
              </div>
              <p className="text-sm text-gray-500 mt-1">
                {new Date(order.created_at).toLocaleDateString()} &middot; {formatPrice(order.total_price)} UZS
              </p>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
