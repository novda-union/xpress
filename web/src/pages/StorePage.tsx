import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { useCartStore } from '../store/cart'
import type { Store, Menu, MenuItem } from '../types'

export default function StorePage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [store, setStore] = useState<Store | null>(null)
  const [menu, setMenu] = useState<Menu | null>(null)
  const [activeCategory, setActiveCategory] = useState<string | null>(null)
  const [selectedItem, setSelectedItem] = useState<MenuItem | null>(null)
  const [selectedModifiers, setSelectedModifiers] = useState<Record<string, string[]>>({})
  const [quantity, setQuantity] = useState(1)

  const cart = useCartStore()

  useEffect(() => {
    if (!slug) return
    api<Store>(`/stores/${slug}`).then(setStore)
    api<Menu>(`/stores/${slug}/menu`).then((m) => {
      setMenu(m)
      if (m.categories.length > 0) setActiveCategory(m.categories[0].id)
    })
  }, [slug])

  useEffect(() => {
    if (store) {
      cart.setStore(store.id, store.slug)
    }
  }, [store])

  const currentCategory = menu?.categories.find((c) => c.id === activeCategory)

  function openItem(item: MenuItem) {
    setSelectedItem(item)
    setSelectedModifiers({})
    setQuantity(1)
    // Pre-select required single-select groups
    item.modifier_groups.forEach((mg) => {
      if (mg.is_required && mg.selection_type === 'single' && mg.modifiers.length > 0) {
        setSelectedModifiers((prev) => ({ ...prev, [mg.id]: [mg.modifiers[0].id] }))
      }
    })
  }

  function toggleModifier(groupId: string, modId: string, type: string) {
    setSelectedModifiers((prev) => {
      const current = prev[groupId] || []
      if (type === 'single') {
        return { ...prev, [groupId]: [modId] }
      }
      // multiple
      if (current.includes(modId)) {
        return { ...prev, [groupId]: current.filter((id) => id !== modId) }
      }
      return { ...prev, [groupId]: [...current, modId] }
    })
  }

  function addToCart() {
    if (!selectedItem) return

    const mods = Object.entries(selectedModifiers).flatMap(([, modIds]) =>
      modIds.map((modId) => {
        const mod = selectedItem.modifier_groups
          .flatMap((mg) => mg.modifiers)
          .find((m) => m.id === modId)
        return mod ? { id: mod.id, name: mod.name, price: mod.price_adjustment } : null
      }).filter(Boolean) as { id: string; name: string; price: number }[]
    )

    const modTotal = mods.reduce((s, m) => s + m.price, 0)
    const itemTotal = (selectedItem.base_price + modTotal) * quantity

    cart.addItem({
      itemId: selectedItem.id,
      name: selectedItem.name,
      price: selectedItem.base_price,
      quantity,
      modifiers: mods,
      totalPrice: itemTotal,
    })

    setSelectedItem(null)
  }

  if (!store || !menu) {
    return <div className="flex items-center justify-center min-h-screen"><p>Loading...</p></div>
  }

  return (
    <div className="max-w-lg mx-auto pb-24">
      {/* Store header */}
      <div className="p-4 bg-white border-b">
        <h1 className="text-xl font-bold">{store.name}</h1>
        <p className="text-sm text-gray-500">{store.description}</p>
        <p className="text-xs text-gray-400 mt-1">{store.address}</p>
      </div>

      {/* Category tabs */}
      <div className="flex overflow-x-auto border-b bg-white sticky top-0 z-10">
        {menu.categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => setActiveCategory(cat.id)}
            className={`px-4 py-3 text-sm whitespace-nowrap border-b-2 ${
              activeCategory === cat.id
                ? 'border-blue-600 text-blue-600 font-medium'
                : 'border-transparent text-gray-500'
            }`}
          >
            {cat.name}
          </button>
        ))}
      </div>

      {/* Items grid */}
      <div className="p-4 space-y-3">
        {currentCategory?.items.map((item) => (
          <button
            key={item.id}
            onClick={() => openItem(item)}
            className="w-full text-left bg-white rounded-lg p-4 shadow-sm border hover:border-blue-300 transition"
          >
            <div className="flex justify-between">
              <div>
                <p className="font-medium">{item.name}</p>
                <p className="text-sm text-gray-500 mt-1">{item.description}</p>
              </div>
              <p className="font-semibold text-nowrap ml-4">{formatPrice(item.base_price)}</p>
            </div>
          </button>
        ))}
      </div>

      {/* Cart floating button */}
      {cart.items.length > 0 && (
        <div className="fixed bottom-0 left-0 right-0 p-4 bg-white border-t">
          <button
            onClick={() => navigate('/cart')}
            className="w-full max-w-lg mx-auto block bg-blue-600 text-white py-3 rounded-lg font-medium"
          >
            Cart ({cart.items.length}) &middot; {formatPrice(cart.total())} UZS
          </button>
        </div>
      )}

      {/* Item detail modal */}
      {selectedItem && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-end">
          <div className="bg-white w-full max-h-[80vh] overflow-y-auto rounded-t-2xl">
            <div className="p-4">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h2 className="text-lg font-bold">{selectedItem.name}</h2>
                  <p className="text-sm text-gray-500">{selectedItem.description}</p>
                </div>
                <button onClick={() => setSelectedItem(null)} className="text-gray-400 text-xl">&times;</button>
              </div>

              {/* Modifier groups */}
              {selectedItem.modifier_groups.map((mg) => (
                <div key={mg.id} className="mb-4">
                  <p className="font-medium text-sm mb-2">
                    {mg.name}
                    {mg.is_required && <span className="text-red-500 ml-1">*</span>}
                  </p>
                  <div className="space-y-2">
                    {mg.modifiers.map((mod) => {
                      const isSelected = (selectedModifiers[mg.id] || []).includes(mod.id)
                      return (
                        <button
                          key={mod.id}
                          onClick={() => toggleModifier(mg.id, mod.id, mg.selection_type)}
                          className={`w-full flex justify-between items-center p-3 rounded border ${
                            isSelected ? 'border-blue-600 bg-blue-50' : 'border-gray-200'
                          }`}
                        >
                          <span className="text-sm">{mod.name}</span>
                          {mod.price_adjustment > 0 && (
                            <span className="text-sm text-gray-500">+{formatPrice(mod.price_adjustment)}</span>
                          )}
                        </button>
                      )
                    })}
                  </div>
                </div>
              ))}

              {/* Quantity */}
              <div className="flex items-center justify-center gap-4 mb-4">
                <button
                  onClick={() => setQuantity(Math.max(1, quantity - 1))}
                  className="w-10 h-10 rounded-full border flex items-center justify-center text-lg"
                >
                  -
                </button>
                <span className="text-lg font-medium w-8 text-center">{quantity}</span>
                <button
                  onClick={() => setQuantity(quantity + 1)}
                  className="w-10 h-10 rounded-full border flex items-center justify-center text-lg"
                >
                  +
                </button>
              </div>

              <button
                onClick={addToCart}
                className="w-full bg-blue-600 text-white py-3 rounded-lg font-medium"
              >
                Add to Cart &middot; {formatPrice(
                  (selectedItem.base_price +
                    Object.entries(selectedModifiers).reduce((sum, [, ids]) => {
                      return sum + ids.reduce((s, id) => {
                        const mod = selectedItem.modifier_groups.flatMap((mg) => mg.modifiers).find((m) => m.id === id)
                        return s + (mod?.price_adjustment || 0)
                      }, 0)
                    }, 0)) * quantity
                )} UZS
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
