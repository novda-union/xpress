import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { BranchCart, CartItem, CartMeta } from '../types'

type AddItem = {
  (item: CartItem): void
  (branchMeta: CartMeta, item: CartItem): void
}

type RemoveItem = {
  (index: number): void
  (branchId: string, index: number): void
}

type UpdateQuantity = {
  (index: number, quantity: number): void
  (branchId: string, index: number, quantity: number): void
}

interface CartStore {
  carts: Record<string, BranchCart>
  activeBranchId: string | null
  branch: CartMeta | null
  items: CartItem[]
  setActiveBranch: (branchId: string) => void
  setBranch: (branch: CartMeta) => void
  addItem: AddItem
  removeItem: RemoveItem
  updateQuantity: UpdateQuantity
  setCartOptions: (branchId: string, options: { paymentMethod?: 'cash' | 'card'; etaMinutes?: number }) => void
  clear: () => void
  clearCart: (branchId: string) => void
  clearAll: () => void
  activeBranchTotal: () => number
  activeBranchCount: () => number
  activeCart: () => BranchCart | null
  totalCartsCount: () => number
  total: () => number
  count: () => number
}

const STORAGE_VERSION = 4

function recalculate(item: CartItem, quantity: number): CartItem {
  const modifierTotal = item.modifiers.reduce((sum, modifier) => sum + modifier.price, 0)
  return {
    ...item,
    quantity,
    totalPrice: (item.price + modifierTotal) * quantity,
  }
}

function getActiveCart(carts: Record<string, BranchCart>, activeBranchId: string | null) {
  if (!activeBranchId) {
    return null
  }
  return carts[activeBranchId] ?? null
}

function withCompatState(carts: Record<string, BranchCart>, activeBranchId: string | null) {
  const activeCart = getActiveCart(carts, activeBranchId)
  return {
    branch: activeCart?.branch ?? null,
    items: activeCart?.items ?? [],
  }
}

function normalizeBranchMeta(
  carts: Record<string, BranchCart>,
  activeBranchId: string | null,
  branchMeta?: CartMeta,
) {
  if (branchMeta) {
    return branchMeta
  }

  const activeCart = getActiveCart(carts, activeBranchId)
  return activeCart?.branch ?? null
}

function removeCartEntry(
  carts: Record<string, BranchCart>,
  branchId: string,
): { carts: Record<string, BranchCart>; nextActiveBranchId: string | null } {
  if (!carts[branchId]) {
    return { carts, nextActiveBranchId: null }
  }

  const nextCarts = { ...carts }
  delete nextCarts[branchId]

  const nextActiveBranchId =
    Object.keys(nextCarts)[0] ?? null

  return {
    carts: nextCarts,
    nextActiveBranchId,
  }
}

export const useCartStore = create<CartStore>()(
  persist(
    (set, get) => ({
      carts: {},
      activeBranchId: null,
      branch: null,
      items: [],

      setActiveBranch: (branchId) => {
        set((state) => ({
          activeBranchId: branchId,
          ...withCompatState(state.carts, branchId),
        }))
      },

      setBranch: (branch) => {
        set((state) => {
          const existing = state.carts[branch.branchId]
          const nextCarts = {
            ...state.carts,
            [branch.branchId]: {
              branch,
              items: existing?.items ?? [],
              paymentMethod: existing?.paymentMethod ?? 'cash',
              etaMinutes: existing?.etaMinutes ?? 15,
            },
          }

          return {
            carts: nextCarts,
            activeBranchId: branch.branchId,
            ...withCompatState(nextCarts, branch.branchId),
          }
        })
      },

      setCartOptions: (branchId, options) => {
        set((state) => {
          const cart = state.carts[branchId]
          if (!cart) return state

          const nextCart: BranchCart = {
            ...cart,
            paymentMethod: options.paymentMethod ?? cart.paymentMethod,
            etaMinutes: options.etaMinutes ?? cart.etaMinutes,
          }

          const nextCarts = { ...state.carts, [branchId]: nextCart }
          return {
            carts: nextCarts,
            ...withCompatState(nextCarts, state.activeBranchId),
          }
        })
      },

      addItem: ((first: CartMeta | CartItem, second?: CartItem) => {
        set((state) => {
          const branchMeta =
            second === undefined
              ? normalizeBranchMeta(state.carts, state.activeBranchId)
              : (first as CartMeta)

          if (!branchMeta) {
            return state
          }

          const item = (second ?? first) as CartItem
          const existing = state.carts[branchMeta.branchId]
          const nextItems = [...(existing?.items ?? []), item]
          const nextCarts = {
            ...state.carts,
            [branchMeta.branchId]: {
              branch: branchMeta,
              items: nextItems,
              paymentMethod: existing?.paymentMethod ?? 'cash',
              etaMinutes: existing?.etaMinutes ?? 15,
            },
          }

          return {
            carts: nextCarts,
            activeBranchId: branchMeta.branchId,
            ...withCompatState(nextCarts, branchMeta.branchId),
          }
        })
      }) as AddItem,

      removeItem: ((first: number | string, second?: number) => {
        set((state) => {
          const branchId =
            typeof first === 'string' ? first : state.activeBranchId
          const index = typeof first === 'string' ? second : first

          if (!branchId || index === undefined) {
            return state
          }

          const cart = state.carts[branchId]
          if (!cart) {
            return state
          }

          const nextItems = cart.items.filter((_, itemIndex) => itemIndex !== index)
          const { carts: nextCarts, nextActiveBranchId } =
            nextItems.length === 0
              ? removeCartEntry(state.carts, branchId)
              : {
                  carts: {
                    ...state.carts,
                    [branchId]: {
                      ...cart,
                      items: nextItems,
                    },
                  },
                  nextActiveBranchId:
                    state.activeBranchId === branchId ? branchId : state.activeBranchId,
                }

          const activeBranchId =
            nextActiveBranchId === branchId && nextItems.length > 0
              ? branchId
              : nextActiveBranchId

          return {
            carts: nextCarts,
            activeBranchId,
            ...withCompatState(nextCarts, activeBranchId),
          }
        })
      }) as RemoveItem,

      updateQuantity: ((first: number | string, second: number, third?: number) => {
        set((state) => {
          const branchId =
            typeof first === 'string' ? first : state.activeBranchId
          const index = typeof first === 'string' ? second : first
          const quantity = typeof first === 'string' ? third : second

          if (!branchId || index === undefined || quantity === undefined) {
            return state
          }

          const cart = state.carts[branchId]
          if (!cart) {
            return state
          }

          const nextItems = cart.items.map((item, itemIndex) =>
            itemIndex === index ? recalculate(item, quantity) : item,
          )
          const nextCarts = {
            ...state.carts,
            [branchId]: {
              ...cart,
              items: nextItems,
            },
          }

          return {
            carts: nextCarts,
            activeBranchId:
              state.activeBranchId ?? branchId,
            ...withCompatState(nextCarts, state.activeBranchId ?? branchId),
          }
        })
      }) as UpdateQuantity,

      clear: () => {
        const activeBranchId = get().activeBranchId
        if (!activeBranchId) {
          return
        }
        get().clearCart(activeBranchId)
      },

      clearCart: (branchId) => {
        set((state) => {
          if (!state.carts[branchId]) {
            return state
          }

          const { carts: nextCarts, nextActiveBranchId } = removeCartEntry(state.carts, branchId)
          const activeBranchId =
            state.activeBranchId === branchId ? nextActiveBranchId : state.activeBranchId

          return {
            carts: nextCarts,
            activeBranchId,
            ...withCompatState(nextCarts, activeBranchId),
          }
        })
      },

      clearAll: () => {
        set({
          carts: {},
          activeBranchId: null,
          branch: null,
          items: [],
        })
      },

      activeBranchTotal: () => {
        const activeCart = getActiveCart(get().carts, get().activeBranchId)
        return activeCart?.items.reduce((sum, item) => sum + item.totalPrice, 0) ?? 0
      },

      activeBranchCount: () => {
        const activeCart = getActiveCart(get().carts, get().activeBranchId)
        return activeCart?.items.reduce((sum, item) => sum + item.quantity, 0) ?? 0
      },

      activeCart: () => getActiveCart(get().carts, get().activeBranchId),

      totalCartsCount: () => Object.keys(get().carts).length,

      total: () => get().activeBranchTotal(),

      count: () => get().activeBranchCount(),
    }),
    {
      name: 'xpressgo-cart-v3',
      version: STORAGE_VERSION,
      migrate: (persistedState, version) => {
        if (version < STORAGE_VERSION) {
          return {
            carts: {},
            activeBranchId: null,
            branch: null,
            items: [],
          }
        }

        const nextState = persistedState as Partial<CartStore>
        const carts = nextState.carts ?? {}
        const activeBranchId = nextState.activeBranchId ?? Object.keys(carts)[0] ?? null

        return {
          ...nextState,
          ...withCompatState(carts, activeBranchId),
        }
      },
    },
  ),
)
