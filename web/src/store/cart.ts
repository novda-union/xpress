import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { CartItem, CartMeta } from '../types'

interface CartState {
  branch: CartMeta | null
  items: CartItem[]
  setBranch: (branch: CartMeta) => void
  addItem: (item: CartItem) => void
  removeItem: (index: number) => void
  updateQuantity: (index: number, quantity: number) => void
  clear: () => void
  total: () => number
  count: () => number
}

function recalculate(item: CartItem, quantity: number): CartItem {
  const modifierTotal = item.modifiers.reduce((sum, modifier) => sum + modifier.price, 0)
  return {
    ...item,
    quantity,
    totalPrice: (item.price + modifierTotal) * quantity,
  }
}

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      branch: null,
      items: [],

      setBranch: (branch) => {
        const current = get().branch
        if (current && current.branchId !== branch.branchId) {
          set({ branch, items: [] })
          return
        }
        set({ branch })
      },

      addItem: (item) =>
        set((state) => ({
          items: [...state.items, item],
        })),

      removeItem: (index) =>
        set((state) => ({
          items: state.items.filter((_, itemIndex) => itemIndex !== index),
        })),

      updateQuantity: (index, quantity) =>
        set((state) => ({
          items: state.items.map((item, itemIndex) =>
            itemIndex === index ? recalculate(item, quantity) : item,
          ),
        })),

      clear: () => set({ items: [] }),

      total: () => get().items.reduce((sum, item) => sum + item.totalPrice, 0),

      count: () => get().items.reduce((sum, item) => sum + item.quantity, 0),
    }),
    { name: 'xpressgo-cart-v2' },
  ),
)
