import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { CartBar } from '../components/cart/CartBar'
import { AppShell } from '../components/layout/AppShell'
import { MenuHeader } from '../components/menu/MenuHeader'
import { ItemCard } from '../components/menu/ItemCard'
import { api } from '../lib/api'
import { useCartStore } from '../store/cart'
import type { BranchDetail, Menu, MenuItem } from '../types'

function BranchPageSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="h-[200px] w-full bg-[var(--xp-card-bg)]" />
      <div className="flex gap-2 px-4 py-3">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-8 w-20 rounded-full bg-[var(--xp-card-bg)]" />
        ))}
      </div>
      <div className="grid grid-cols-2 gap-3 px-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="aspect-square rounded-xl bg-[var(--xp-card-bg)]" />
        ))}
      </div>
    </div>
  )
}

export default function BranchPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const cart = useCartStore()
  const [detail, setDetail] = useState<BranchDetail | null>(null)
  const [menu, setMenu] = useState<Menu | null>(null)
  const [activeCategory, setActiveCategory] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return

    api<BranchDetail>(`/branches/${id}`).then((nextDetail) => {
      setDetail(nextDetail)
      cart.setBranch({
        branchId: nextDetail.branch.id,
        branchName: nextDetail.branch.name,
        storeName: nextDetail.store.name,
        bannerImageUrl: nextDetail.branch.banner_image_url,
      })
    })

    api<Menu>(`/branches/${id}/menu`).then((nextMenu) => {
      setMenu(nextMenu)
      if (nextMenu.categories.length > 0) {
        setActiveCategory(nextMenu.categories[0].id)
      }
    })
  }, [id]) // eslint-disable-line react-hooks/exhaustive-deps

  const currentCategory = useMemo(() => {
    if (!menu) return null
    if (activeCategory) {
      return menu.categories.find((c) => c.id === activeCategory) ?? menu.categories[0] ?? null
    }
    return menu.categories[0] ?? null
  }, [activeCategory, menu])

  const cartCount = cart.activeBranchCount()
  const cartTotal = cart.activeBranchTotal()

  return (
    <AppShell
      header={<MenuHeader title={detail?.store.name ?? ''} count={cartCount} />}
      bottomBar={
        cartCount > 0 ? (
          <CartBar count={cartCount} total={cartTotal} totalCartsCount={cart.totalCartsCount()} onOpen={() => navigate('/cart')} />
        ) : null
      }
    >
      {!detail || !menu ? (
        <BranchPageSkeleton />
      ) : (
        <>
          <div className="relative">
            <img
              src={detail.branch.banner_image_url || 'https://placehold.co/900x400?text=Xpressgo'}
              alt={detail.branch.name}
              className="h-[200px] w-full object-cover"
            />
            <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/70 to-transparent px-4 py-6">
              <p className="text-[22px] font-bold text-white">{detail.store.name}</p>
              <p className="text-sm text-white/80">{detail.branch.name}</p>
            </div>
          </div>

          <div className="sticky top-12 z-10 bg-[var(--tg-theme-bg-color)]">
            <div className="scrollbar-none flex gap-2 overflow-x-auto px-4 py-3">
              {menu.categories.map((category) => (
                <button
                  type="button"
                  key={category.id}
                  onClick={() => setActiveCategory(category.id)}
                  className={`xp-pill whitespace-nowrap px-4 text-[13px] font-medium ${
                    activeCategory === category.id
                      ? 'bg-[var(--xp-brand)] text-white'
                      : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
                  }`}
                >
                  {category.name}
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 px-4 pb-28">
            {(currentCategory?.items ?? []).map((item: MenuItem) => (
              <ItemCard
                key={item.id}
                item={item}
                onSelect={(selected) => navigate(`/item/${selected.id}?branch=${detail.branch.id}`)}
              />
            ))}
          </div>
        </>
      )}
    </AppShell>
  )
}
