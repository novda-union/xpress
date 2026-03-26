import { useEffect, useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { useWebSocket } from '../hooks/useWebSocket'
import type { Order } from '../types'

const STATUS_STEPS = ['pending', 'accepted', 'preparing', 'ready', 'picked_up']
const STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  accepted: 'Accepted',
  preparing: 'Preparing',
  ready: 'Ready for Pickup',
  picked_up: 'Picked Up',
  rejected: 'Rejected',
  cancelled: 'Cancelled',
}

export default function OrderPage() {
  const { id } = useParams<{ id: string }>()
  const [order, setOrder] = useState<Order | null>(null)

  useEffect(() => {
    if (!id) return
    api<Order>(`/orders/${id}`).then(setOrder)
  }, [id])

  const handleWsMessage = useCallback(
    (msg: any) => {
      if (msg.order_id === id) {
        // Refresh order on status change
        api<Order>(`/orders/${id}`).then(setOrder)
      }
    },
    [id]
  )

  useWebSocket(handleWsMessage)

  if (!order) {
    return <div className="flex items-center justify-center min-h-screen"><p>Loading...</p></div>
  }

  const currentStep = STATUS_STEPS.indexOf(order.status)
  const isTerminal = ['picked_up', 'rejected', 'cancelled'].includes(order.status)

  return (
    <div className="max-w-lg mx-auto p-4">
      <h1 className="text-xl font-bold mb-2">Order #{order.order_number}</h1>

      {/* Status */}
      <div className="bg-white rounded-lg p-4 shadow-sm border mb-4">
        <p className="text-lg font-semibold mb-4">{STATUS_LABELS[order.status]}</p>

        {!isTerminal && (
          <div className="flex items-center gap-1 mb-4">
            {STATUS_STEPS.slice(0, -1).map((step, i) => (
              <div
                key={step}
                className={`flex-1 h-2 rounded ${
                  i <= currentStep ? 'bg-blue-600' : 'bg-gray-200'
                }`}
              />
            ))}
          </div>
        )}

        {order.status === 'rejected' && order.rejection_reason && (
          <p className="text-red-600 text-sm">Reason: {order.rejection_reason}</p>
        )}
      </div>

      {/* Items */}
      <div className="bg-white rounded-lg p-4 shadow-sm border mb-4">
        <h2 className="font-medium mb-3">Items</h2>
        <div className="space-y-2">
          {order.items.map((item) => (
            <div key={item.id} className="flex justify-between text-sm">
              <div>
                <span>{item.quantity}x {item.item_name}</span>
                {item.modifiers.length > 0 && (
                  <p className="text-gray-500 text-xs">
                    {item.modifiers.map((m) => m.modifier_name).join(', ')}
                  </p>
                )}
              </div>
              <span>{formatPrice(item.item_price * item.quantity)}</span>
            </div>
          ))}
        </div>
        <div className="border-t mt-3 pt-3 flex justify-between font-bold">
          <span>Total</span>
          <span>{formatPrice(order.total_price)} UZS</span>
        </div>
      </div>

      {/* ETA */}
      <div className="bg-white rounded-lg p-4 shadow-sm border">
        <p className="text-sm text-gray-500">Estimated arrival: ~{order.eta_minutes} min</p>
        <p className="text-sm text-gray-500">Payment: {order.payment_method === 'pay_at_pickup' ? 'Pay at pickup' : 'Paid'}</p>
      </div>
    </div>
  )
}
