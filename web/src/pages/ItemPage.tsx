import { ArrowLeft, Minus, Plus, ShoppingCart } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { ModifierGroupSelector } from '../components/menu/ModifierGroupSelector'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { useCartStore } from '../store/cart'
import type { ItemDetailResponse, ModifierGroup } from '../types'

function ItemPageSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="aspect-[4/3] w-full bg-[var(--xp-card-bg)]" />
      <div className="space-y-3 px-4 pt-4">
        <div className="h-6 w-3/4 rounded-lg bg-[var(--xp-card-bg)]" />
        <div className="h-5 w-1/3 rounded-lg bg-[var(--xp-card-bg)]" />
        <div className="space-y-2">
          <div className="h-4 w-full rounded bg-[var(--xp-card-bg)]" />
          <div className="h-4 w-5/6 rounded bg-[var(--xp-card-bg)]" />
        </div>
      </div>
    </div>
  )
}

function buildInitialSelections(modifierGroups: ItemDetailResponse['item']['modifier_groups']) {
  const nextSelections: Record<string, string[]> = {}
  modifierGroups.forEach((group) => {
    if (group.is_required && group.selection_type === 'single' && group.modifiers[0]) {
      nextSelections[group.id] = [group.modifiers[0].id]
    }
  })
  return nextSelections
}

export default function ItemPage() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const branchId = searchParams.get('branch')
  const navigate = useNavigate()
  const cart = useCartStore()

  const [data, setData] = useState<ItemDetailResponse | null>(null)
  const [selectedModifiers, setSelectedModifiers] = useState<Record<string, string[]>>({})
  const [quantity, setQuantity] = useState(1)
  const [stickyVisible, setStickyVisible] = useState(false)
  const heroRef = useRef<HTMLImageElement>(null)

  useEffect(() => {
    let active = true

    if (!branchId || !id) {
      return () => {
        active = false
      }
    }

    api<ItemDetailResponse>(`/items/${id}?branch=${branchId}`).then((nextData) => {
      if (!active) {
        return
      }

      setData(nextData)
      setSelectedModifiers(buildInitialSelections(nextData.item.modifier_groups))
      setQuantity(1)
      setStickyVisible(false)
    })

    return () => {
      active = false
    }
  }, [branchId, id])

  useEffect(() => {
    const hero = heroRef.current
    if (!hero) {
      return undefined
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        setStickyVisible(!entry.isIntersecting)
      },
      { threshold: 0 },
    )

    observer.observe(hero)
    return () => observer.disconnect()
  }, [data])

  const total = useMemo(() => {
    if (!data) {
      return 0
    }

    const modifierTotal = Object.entries(selectedModifiers).reduce((sum, [, ids]) => {
      return (
        sum +
        ids.reduce((innerSum, modifierId) => {
          const modifier = data.item.modifier_groups
            .flatMap((group) => group.modifiers)
            .find((entry) => entry.id === modifierId)
          return innerSum + (modifier?.price_adjustment ?? 0)
        }, 0)
      )
    }, 0)

    return (data.item.base_price + modifierTotal) * quantity
  }, [data, quantity, selectedModifiers])

  function toggleModifier(group: ModifierGroup, modifierId: string) {
    setSelectedModifiers((current) => {
      const selected = current[group.id] ?? []

      if (group.selection_type === 'single') {
        return { ...current, [group.id]: [modifierId] }
      }

      return {
        ...current,
        [group.id]: selected.includes(modifierId)
          ? selected.filter((entry) => entry !== modifierId)
          : [...selected, modifierId],
      }
    })
  }

  function addToCart() {
    if (!data || !branchId) {
      return
    }

    const branchMeta = {
      branchId: data.branch.branch.id,
      branchName: data.branch.branch.name,
      storeName: data.branch.store.name,
      bannerImageUrl: data.branch.branch.banner_image_url,
    }

    const modifiers = Object.entries(selectedModifiers).flatMap(([, ids]) =>
      ids
        .map((modifierId) =>
          data.item.modifier_groups
            .flatMap((group) => group.modifiers)
            .find((modifier) => modifier.id === modifierId),
        )
        .filter(Boolean)
        .map((modifier) => ({
          id: modifier!.id,
          name: modifier!.name,
          price: modifier!.price_adjustment,
        })),
    )

    cart.addItem(branchMeta, {
      itemId: data.item.id,
      imageUrl: data.item.image_url,
      name: data.item.name,
      price: data.item.base_price,
      quantity,
      modifiers,
      totalPrice: total,
    })

    navigate('/cart')
  }

  const cartCount = cart.activeBranchCount()
  const isReady =
    Boolean(data) && data?.item.id === id && data?.branch.branch.id === branchId

  return (
    <div className="min-h-dvh bg-[var(--tg-theme-bg-color)]">
      <div
        className="fixed inset-x-0 top-0 z-10 flex items-center gap-3 border-b border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)]/90 px-4 py-3 backdrop-blur-sm"
        style={{
          opacity: stickyVisible ? 1 : 0,
          pointerEvents: stickyVisible ? 'auto' : 'none',
          transition: 'opacity 150ms ease',
        }}
      >
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="flex h-9 w-9 items-center justify-center rounded-full bg-[var(--xp-card-bg)]"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        <p className="flex-1 truncate text-[15px] font-semibold">{data?.item.name ?? ''}</p>
        <button
          type="button"
          onClick={() => navigate('/cart')}
          className="relative flex h-9 w-9 items-center justify-center rounded-full bg-[var(--xp-card-bg)]"
        >
          <ShoppingCart className="h-4 w-4" />
          {cartCount > 0 && (
            <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-[var(--xp-brand)] text-[9px] font-bold text-white">
              {cartCount}
            </span>
          )}
        </button>
      </div>

      {!isReady || !data ? (
        <ItemPageSkeleton />
      ) : (
        <>
          <div className="relative">
            <img
              ref={heroRef}
              src={data.item.image_url || 'https://placehold.co/800x600?text=Item'}
              alt={data.item.name}
              className="aspect-[4/3] w-full object-cover"
            />
            <div className="absolute inset-x-0 bottom-0 h-20 bg-gradient-to-t from-[var(--tg-theme-bg-color)] to-transparent" />
            <button
              type="button"
              onClick={() => navigate(-1)}
              className="absolute left-4 top-4 z-30 flex h-11 w-11 items-center justify-center rounded-full bg-black/40 text-white backdrop-blur-sm"
            >
              <ArrowLeft className="h-5 w-5" />
            </button>
          </div>

          <div className="px-4 pb-32 pt-4">
            <h1 className="text-[22px] font-bold">{data.item.name}</h1>
            <p className="mt-1 text-xl font-semibold text-[var(--xp-brand)]">
              {formatPrice(data.item.base_price)} UZS
            </p>
            {data.item.description ? (
              <p className="mt-3 text-[15px] leading-6 text-[var(--tg-theme-hint-color)]">
                {data.item.description}
              </p>
            ) : null}

            {data.item.modifier_groups.map((group) => (
              <ModifierGroupSelector
                key={group.id}
                group={group}
                selected={selectedModifiers[group.id] ?? []}
                onToggle={toggleModifier}
              />
            ))}
          </div>
        </>
      )}

      {isReady && data ? (
        <div className="fixed inset-x-0 bottom-0 z-20 border-t border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-3">
          <div className="mx-auto flex max-w-[32rem] items-center gap-4">
            <div className="flex items-center gap-2 rounded-[20px] bg-[var(--xp-card-bg)] px-3 py-2">
              <button
                type="button"
                onClick={() => setQuantity((value) => Math.max(1, value - 1))}
                className="flex h-9 w-9 items-center justify-center rounded-full"
              >
                <Minus className="h-4 w-4" />
              </button>
              <span className="w-6 text-center text-xl font-semibold">{quantity}</span>
              <button
                type="button"
                onClick={() => setQuantity((value) => value + 1)}
                className="flex h-9 w-9 items-center justify-center rounded-full"
              >
                <Plus className="h-4 w-4" />
              </button>
            </div>

            <button
              type="button"
              onClick={addToCart}
              className="flex h-[52px] flex-1 items-center justify-center rounded-[20px] bg-[var(--xp-brand)] px-4 text-sm font-semibold text-white"
            >
              Add to Cart · {formatPrice(total)} UZS
            </button>
          </div>
        </div>
      ) : null}
    </div>
  )
}
