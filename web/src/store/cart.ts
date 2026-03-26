import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { CartItem } from '../types'

interface CartState {
  storeId: string | null
  storeSlug: string | null
  items: CartItem[]
  setStore: (id: string, slug: string) => void
  addItem: (item: CartItem) => void
  removeItem: (index: number) => void
  updateQuantity: (index: number, quantity: number) => void
  clear: () => void
  total: () => number
}

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      storeId: null,
      storeSlug: null,
      items: [],

      setStore: (id, slug) => {
        const state = get()
        // Clear cart if switching stores
        if (state.storeId && state.storeId !== id) {
          set({ storeId: id, storeSlug: slug, items: [] })
        } else {
          set({ storeId: id, storeSlug: slug })
        }
      },

      addItem: (item) => set((state) => ({ items: [...state.items, item] })),

      removeItem: (index) =>
        set((state) => ({ items: state.items.filter((_, i) => i !== index) })),

      updateQuantity: (index, quantity) =>
        set((state) => ({
          items: state.items.map((item, i) =>
            i === index ? { ...item, quantity, totalPrice: (item.price + item.modifiers.reduce((s, m) => s + m.price, 0)) * quantity } : item
          ),
        })),

      clear: () => set({ items: [] }),

      total: () => get().items.reduce((sum, item) => sum + item.totalPrice, 0),
    }),
    { name: 'xpressgo-cart' }
  )
)
