import { Minus, Plus } from 'lucide-react'
import { useState } from 'react'
import type { MouseEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { BranchConflictSheet } from '../cart/BranchConflictSheet'
import { calculateDistanceInKm, formatDistance } from '../../lib/distance'
import { formatPrice } from '../../lib/format'
import { useCartStore } from '../../store/cart'
import type { CartItem, DiscoverItem } from '../../types'

interface DiscoverItemCardProps {
  item: DiscoverItem
  userLat: number
  userLng: number
}

function isNewItem(createdAt: string): boolean {
  const created = new Date(createdAt)
  const sevenDaysAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)
  return created > sevenDaysAgo
}

function findLastItemIndex(items: CartItem[], itemId: string): number {
  for (let index = items.length - 1; index >= 0; index -= 1) {
    if (items[index].itemId === itemId) {
      return index
    }
  }
  return -1
}

export function DiscoverItemCard({ item, userLat, userLng }: DiscoverItemCardProps) {
  const navigate = useNavigate()
  const cart = useCartStore()
  const [showConflict, setShowConflict] = useState(false)

  const branchMeta = {
    branchId: item.branch_id,
    branchName: item.branch_name,
    storeName: item.store_name,
    bannerImageUrl: '',
  }

  const activeCart = cart.activeCart()
  const branchCart = cart.carts[item.branch_id]
  const existingItems = branchCart?.items ?? []
  const itemCount = existingItems.reduce((sum, cartItem) => {
    return cartItem.itemId === item.id ? sum + cartItem.quantity : sum
  }, 0)
  const hasAnyCart = Object.keys(cart.carts).length > 0
  const distanceKm = calculateDistanceInKm(userLat, userLng, item.lat, item.lng)
  const canAddDirectly =
    item.is_available &&
    (!hasAnyCart || activeCart?.branch.branchId === item.branch_id || Boolean(branchCart))

  function buildCartItem(): CartItem {
    return {
      itemId: item.id,
      imageUrl: item.image_url,
      name: item.name,
      price: item.base_price,
      quantity: 1,
      modifiers: [],
      totalPrice: item.base_price,
    }
  }

  function handleCardClick() {
    navigate(`/item/${item.id}?branch=${item.branch_id}`)
  }

  function handleAdd(event?: MouseEvent<HTMLButtonElement>) {
    event?.stopPropagation()

    if (!item.is_available) {
      return
    }

    if (item.has_required_modifiers) {
      navigate(`/item/${item.id}?branch=${item.branch_id}`)
      return
    }

    if (canAddDirectly) {
      cart.addItem(branchMeta, buildCartItem())
      return
    }

    setShowConflict(true)
  }

  function handleConfirmConflict() {
    setShowConflict(false)
    cart.addItem(branchMeta, buildCartItem())
  }

  function handleDecrement(event: MouseEvent<HTMLButtonElement>) {
    event.stopPropagation()

    const lastIndex = findLastItemIndex(existingItems, item.id)
    if (lastIndex === -1) {
      return
    }

    const target = existingItems[lastIndex]
    if (target.quantity > 1) {
      cart.updateQuantity(item.branch_id, lastIndex, target.quantity - 1)
      return
    }

    cart.removeItem(item.branch_id, lastIndex)
  }

  const showStepper = itemCount > 0 && !item.has_required_modifiers

  return (
    <>
      <div
        role="button"
        tabIndex={0}
        onClick={handleCardClick}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault()
            handleCardClick()
          }
        }}
        className="xp-card cursor-pointer overflow-hidden text-left"
      >
        <div className="relative">
          <img
            src={item.image_url || 'https://placehold.co/300x300?text=Item'}
            alt={item.name}
            loading="lazy"
            className={`h-36 w-full object-cover ${item.is_available ? '' : 'grayscale'}`}
          />
          {isNewItem(item.created_at) ? (
            <span className="absolute left-2 top-2 rounded-full bg-emerald-500 px-2 py-0.5 text-[10px] font-semibold text-white">
              NEW
            </span>
          ) : null}
          {!item.is_available ? (
            <span className="absolute right-2 top-2 rounded-full bg-black/70 px-2 py-1 text-[11px] font-semibold text-white">
              Unavailable
            </span>
          ) : null}
          <div
            className="absolute bottom-2 right-2"
            onClick={(event) => event.stopPropagation()}
          >
            {showStepper ? (
              <div
                className="flex items-center overflow-hidden rounded-full bg-[var(--xp-brand)]"
                style={{ width: '88px', transition: 'width 200ms ease' }}
              >
                <button
                  type="button"
                  onClick={handleDecrement}
                  className="flex h-9 w-9 shrink-0 items-center justify-center text-white"
                  aria-label="Remove one"
                >
                  <Minus className="h-3.5 w-3.5" />
                </button>
                <span className="flex-1 text-center text-sm font-semibold text-white">
                  {itemCount}
                </span>
                <button
                  type="button"
                  onClick={handleAdd}
                  disabled={!item.is_available}
                  className="flex h-9 w-9 shrink-0 items-center justify-center text-white disabled:opacity-50"
                  aria-label="Add one more"
                >
                  <Plus className="h-3.5 w-3.5" />
                </button>
              </div>
            ) : (
              <button
                type="button"
                onClick={handleAdd}
                disabled={!item.is_available}
                className="flex h-9 w-9 items-center justify-center rounded-full bg-[var(--xp-brand)] text-white transition-transform duration-150 active:scale-110 disabled:opacity-50"
                aria-label={`Add ${item.name} to cart`}
              >
                <Plus className="h-4 w-4" />
              </button>
            )}
          </div>
        </div>

        <div className="p-2 pb-3">
          <p className="line-clamp-2 text-[13px] font-semibold leading-5">{item.name}</p>
          <p className="mt-0.5 text-[13px] font-semibold text-[var(--xp-brand)]">
            {formatPrice(item.base_price)} UZS
          </p>
          <p className="mt-0.5 truncate text-[11px] text-[var(--tg-theme-hint-color)]">
            {item.store_name} · {formatDistance(distanceKm)}
          </p>
        </div>
      </div>

      {showConflict ? (
        <BranchConflictSheet
          newBranchName={item.branch_name}
          onConfirm={handleConfirmConflict}
          onCancel={() => setShowConflict(false)}
        />
      ) : null}
    </>
  )
}
