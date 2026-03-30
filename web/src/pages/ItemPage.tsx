import { Minus, Plus } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { MenuHeader } from '../components/menu/MenuHeader'
import { ModifierGroupSelector } from '../components/menu/ModifierGroupSelector'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { AppShell } from '../components/layout/AppShell'
import { useCartStore } from '../store/cart'
import type { BranchDetail, Menu, MenuItem, ModifierGroup } from '../types'

export default function ItemPage() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const branchId = searchParams.get('branch')
  const cart = useCartStore()
  const navigate = useNavigate()

  const [detail, setDetail] = useState<BranchDetail | null>(null)
  const [item, setItem] = useState<MenuItem | null>(null)
  const [selectedModifiers, setSelectedModifiers] = useState<Record<string, string[]>>({})
  const [quantity, setQuantity] = useState(1)

  useEffect(() => {
    if (!branchId || !id) {
      return
    }

    api<BranchDetail>(`/branches/${branchId}`).then(setDetail)
    api<Menu>(`/branches/${branchId}/menu`).then((menu) => {
      const nextItem = menu.categories.flatMap((category) => category.items).find((menuItem) => menuItem.id === id) ?? null
      setItem(nextItem)
      if (nextItem) {
        const nextSelections: Record<string, string[]> = {}
        nextItem.modifier_groups.forEach((group) => {
          if (group.is_required && group.selection_type === 'single' && group.modifiers[0]) {
            nextSelections[group.id] = [group.modifiers[0].id]
          }
        })
        setSelectedModifiers(nextSelections)
      }
    })
  }, [branchId, id])

  const total = useMemo(() => {
    if (!item) {
      return 0
    }
    const modifierTotal = Object.entries(selectedModifiers).reduce((sum, [, ids]) => {
      return sum + ids.reduce((innerSum, modifierId) => {
        const modifier = item.modifier_groups.flatMap((group) => group.modifiers).find((entry) => entry.id === modifierId)
        return innerSum + (modifier?.price_adjustment ?? 0)
      }, 0)
    }, 0)
    return (item.base_price + modifierTotal) * quantity
  }, [item, quantity, selectedModifiers])

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
    if (!item || !detail) {
      return
    }

    cart.setBranch({
      branchId: detail.branch.id,
      branchName: detail.branch.name,
      storeName: detail.store.name,
      bannerImageUrl: detail.branch.banner_image_url,
    })

    const modifiers = Object.entries(selectedModifiers).flatMap(([, ids]) =>
      ids
        .map((modifierId) =>
          item.modifier_groups
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

    cart.addItem({
      itemId: item.id,
      imageUrl: item.image_url,
      name: item.name,
      price: item.base_price,
      quantity,
      modifiers,
      totalPrice: total,
    })

    navigate('/cart')
  }

  if (!item || !detail) {
    return <div className="min-h-screen flex items-center justify-center">Loading...</div>
  }

  return (
    <AppShell header={<MenuHeader title={detail.store.name} count={cart.count()} />}>
      <img
        src={item.image_url || 'https://placehold.co/1000x600?text=Item'}
        alt={item.name}
        className="h-[280px] w-full object-cover"
      />
      <div className="px-4 pb-28 pt-4">
        <h1 className="text-[22px] font-bold">{item.name}</h1>
        <p className="mt-2 text-xl font-semibold text-[var(--xp-brand)]">{formatPrice(item.base_price)} UZS</p>
        <p className="mt-3 text-[15px] leading-6 text-[var(--tg-theme-hint-color)]">{item.description}</p>

        {item.modifier_groups.map((group) => (
          <ModifierGroupSelector
            key={group.id}
            group={group}
            selected={selectedModifiers[group.id] ?? []}
            onToggle={toggleModifier}
          />
        ))}
      </div>

      <div className="fixed inset-x-0 bottom-0 z-20 border-t border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-3">
        <div className="mx-auto flex max-w-[32rem] items-center gap-4">
          <div className="flex items-center gap-2 rounded-[20px] bg-[var(--xp-card-bg)] px-3 py-2">
            <button type="button" className="flex h-9 w-9 items-center justify-center rounded-full" onClick={() => setQuantity((value) => Math.max(1, value - 1))}>
              <Minus className="h-4 w-4" />
            </button>
            <span className="w-6 text-center text-xl font-semibold">{quantity}</span>
            <button type="button" className="flex h-9 w-9 items-center justify-center rounded-full" onClick={() => setQuantity((value) => value + 1)}>
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
    </AppShell>
  )
}
