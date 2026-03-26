import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { useCartStore } from '../store/cart'
import type { Order } from '../types'

const ETA_OPTIONS = [5, 10, 15, 20, 30]

export default function CartPage() {
  const navigate = useNavigate()
  const cart = useCartStore()
  const [eta, setEta] = useState(15)
  const [loading, setLoading] = useState(false)

  async function placeOrder() {
    if (cart.items.length === 0 || !cart.storeId) return
    setLoading(true)

    try {
      const order = await api<Order>('/orders', {
        method: 'POST',
        body: JSON.stringify({
          store_id: cart.storeId,
          payment_method: 'pay_at_pickup',
          eta_minutes: eta,
          items: cart.items.map((item) => ({
            item_id: item.itemId,
            item_name: item.name,
            item_price: item.price,
            quantity: item.quantity,
            modifiers: item.modifiers.map((m) => ({
              modifier_id: m.id,
              modifier_name: m.name,
              price_adjustment: m.price,
            })),
          })),
        }),
      })

      cart.clear()
      navigate(`/order/${order.id}`)
    } catch (e: any) {
      alert(e.message || 'Failed to place order')
    } finally {
      setLoading(false)
    }
  }

  if (cart.items.length === 0) {
    return (
      <div className="max-w-lg mx-auto p-4 text-center mt-20">
        <p className="text-gray-500 mb-4">Your cart is empty</p>
        <button
          onClick={() => navigate(`/${cart.storeSlug || ''}`)}
          className="text-blue-600 font-medium"
        >
          Browse Menu
        </button>
      </div>
    )
  }

  return (
    <div className="max-w-lg mx-auto pb-32">
      <div className="p-4 bg-white border-b">
        <div className="flex items-center gap-3">
          <button onClick={() => navigate(-1)} className="text-gray-500">&larr;</button>
          <h1 className="text-xl font-bold">Your Cart</h1>
        </div>
      </div>

      {/* Cart items */}
      <div className="p-4 space-y-3">
        {cart.items.map((item, index) => (
          <div key={index} className="bg-white rounded-lg p-4 shadow-sm border">
            <div className="flex justify-between">
              <div>
                <p className="font-medium">{item.quantity}x {item.name}</p>
                {item.modifiers.length > 0 && (
                  <p className="text-sm text-gray-500">
                    {item.modifiers.map((m) => m.name).join(', ')}
                  </p>
                )}
              </div>
              <div className="flex items-start gap-2">
                <span className="font-semibold">{formatPrice(item.totalPrice)}</span>
                <button
                  onClick={() => cart.removeItem(index)}
                  className="text-red-400 text-sm ml-2"
                >
                  &times;
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* ETA selector */}
      <div className="p-4">
        <p className="font-medium mb-3">I'll arrive in:</p>
        <div className="flex gap-2 flex-wrap">
          {ETA_OPTIONS.map((minutes) => (
            <button
              key={minutes}
              onClick={() => setEta(minutes)}
              className={`px-4 py-2 rounded-full border text-sm ${
                eta === minutes
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'border-gray-300 text-gray-600'
              }`}
            >
              {minutes} min
            </button>
          ))}
        </div>
      </div>

      {/* Place order */}
      <div className="fixed bottom-0 left-0 right-0 p-4 bg-white border-t">
        <button
          onClick={placeOrder}
          disabled={loading}
          className="w-full max-w-lg mx-auto block bg-blue-600 text-white py-3 rounded-lg font-medium disabled:opacity-50"
        >
          {loading ? 'Placing order...' : `Place Order · ${formatPrice(cart.total())} UZS`}
        </button>
        <p className="text-center text-xs text-gray-400 mt-2">Pay at pickup</p>
      </div>
    </div>
  )
}
