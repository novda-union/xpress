import { lazy, Suspense, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { ShoppingBag } from 'lucide-react'
import { BranchSheet } from '../components/discovery/BranchSheet'
import { ViewToggle } from '../components/discovery/ViewToggle'
import { DiscoverItemCard } from '../components/discovery/DiscoverItemCard'
import { CartBar } from '../components/cart/CartBar'
import { useDiscoveryFeed } from '../hooks/useDiscoveryFeed'
import { useDiscoveryItems } from '../hooks/useDiscoveryItems'
import { useDiscovery } from '../hooks/useDiscovery'
import { useUserLocation } from '../hooks/useUserLocation'
import { useTelegramTheme } from '../hooks/useTelegramTheme'
import { useCartStore } from '../store/cart'
import type { DiscoverBranch } from '../types'

const DiscoveryMap = lazy(() =>
  import('../components/discovery/DiscoveryMap').then((module) => ({ default: module.DiscoveryMap })),
)

type ChipId = 'new' | 'popular' | 'bar' | 'cafe' | 'coffee' | 'restaurant' | 'fastfood'

const CHIPS: Array<{ id: ChipId; label: string }> = [
  { id: 'new', label: 'New' },
  { id: 'popular', label: 'Popular' },
  { id: 'bar', label: 'Bars' },
  { id: 'cafe', label: 'Cafes' },
  { id: 'coffee', label: 'Coffee' },
  { id: 'restaurant', label: 'Restaurants' },
  { id: 'fastfood', label: 'Fast Food' },
]

function getSortAndCategory(chip: ChipId): { sort: 'new' | 'popular'; category: string } {
  if (chip === 'popular') {
    return { sort: 'popular', category: '' }
  }
  if (chip === 'new') {
    return { sort: 'new', category: '' }
  }
  return { sort: 'new', category: chip }
}

function FeedSectionSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="mb-3 flex items-center justify-between px-4">
        <div className="h-5 w-28 rounded bg-[var(--xp-card-bg)]" />
        <div className="h-4 w-12 rounded bg-[var(--xp-card-bg)]" />
      </div>
      <div className="flex gap-3 overflow-hidden px-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="w-[160px] shrink-0">
            <div className="aspect-square w-full rounded-xl bg-[var(--xp-card-bg)]" />
            <div className="mt-2 h-4 w-3/4 rounded bg-[var(--xp-card-bg)]" />
            <div className="mt-1 h-3 w-1/2 rounded bg-[var(--xp-card-bg)]" />
          </div>
        ))}
      </div>
    </div>
  )
}

function GridSkeleton() {
  return (
    <div className="grid animate-pulse grid-cols-2 gap-3">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="aspect-square rounded-xl bg-[var(--xp-card-bg)]" />
      ))}
    </div>
  )
}

export default function HomePage() {
  useTelegramTheme()

  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const storeSlug = searchParams.get('store')
  const location = useUserLocation()
  const cart = useCartStore()
  const [activeChip, setActiveChip] = useState<ChipId>('new')
  const [view, setView] = useState<'list' | 'map'>('list')
  const [selectedBranch, setSelectedBranch] = useState<(DiscoverBranch & { distanceKm: number }) | null>(null)
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  const { sort, category } = getSortAndCategory(activeChip)
  const { sections, loading: feedLoading } = useDiscoveryFeed()
  const {
    items,
    loading: itemsLoading,
    loadingMore,
    hasMore,
    loadMore,
  } = useDiscoveryItems(category, sort)
  const discovery = useDiscovery(location.lat, location.lng)
  const visibleBranches = useMemo(
    () =>
      discovery.branches.filter((branch) => !storeSlug || branch.store_slug === storeSlug),
    [discovery.branches, storeSlug],
  )

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel || view !== 'list') {
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          loadMore()
        }
      },
      { rootMargin: '200px' },
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore, view])

  const cartCount = cart.activeBranchCount()
  const cartTotal = cart.activeBranchTotal()
  const hour = new Date().getHours()
  const greeting =
    hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening'

  const gridHeading = useMemo(() => {
    if (activeChip === 'new') {
      return 'All Items'
    }
    if (activeChip === 'popular') {
      return 'All Popular'
    }
    return CHIPS.find((chip) => chip.id === activeChip)?.label ?? 'All Items'
  }, [activeChip])

  return (
    <div className="min-h-dvh bg-[var(--tg-theme-bg-color)]">
      {view === 'map' ? (
        <div className="fixed inset-0 z-0 px-4 pb-24 pt-4">
          <Suspense
            fallback={
              <div className="h-[calc(100vh-5rem)] w-full animate-pulse rounded-[24px] bg-[var(--xp-card-bg)]" />
            }
          >
            <DiscoveryMap
              branches={visibleBranches}
              center={location}
              selectedBranchId={selectedBranch?.branch_id ?? null}
              visible
              onSelect={setSelectedBranch}
            />
          </Suspense>
        </div>
      ) : null}

      {view === 'list' ? (
        <div className="flex flex-col pb-28">
          <div className="flex items-start justify-between px-4 pb-4 pt-6">
            <div>
              <h1 className="text-[22px] font-bold">{greeting}</h1>
              <p className="mt-1 text-[14px] text-[var(--tg-theme-hint-color)]">
                What are you craving?
              </p>
            </div>
            <button
              type="button"
              onClick={() => navigate('/cart')}
              className="relative flex h-10 w-10 items-center justify-center rounded-full bg-[var(--xp-card-bg)]"
            >
              <ShoppingBag size={22} />
              {cart.totalCartsCount() > 0 ? (
                <span className="absolute right-0 top-0 h-2.5 w-2.5 rounded-full bg-[var(--xp-brand)]" />
              ) : null}
            </button>
          </div>

          <div className="sticky top-0 z-10 bg-[var(--tg-theme-bg-color)]">
            <div className="scrollbar-none flex gap-2 overflow-x-auto px-4 py-2">
              {CHIPS.map((chip) => (
                <button
                  key={chip.id}
                  type="button"
                  onClick={() => setActiveChip(chip.id)}
                  className={`xp-pill shrink-0 whitespace-nowrap px-4 text-[13px] font-medium ${
                    activeChip === chip.id
                      ? 'bg-[var(--xp-brand)] text-white'
                      : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
                  }`}
                >
                  {chip.label}
                </button>
              ))}
            </div>
          </div>

          {activeChip === 'new' || activeChip === 'popular' ? (
            <div className="mt-2 space-y-6">
              {feedLoading
                ? [1, 2].map((i) => <FeedSectionSkeleton key={i} />)
                : sections.map((section) => (
                    <div key={section.type}>
                      <div className="mb-3 flex items-center justify-between px-4">
                        <h2 className="text-[15px] font-semibold">{section.title}</h2>
                        <button
                          type="button"
                          className="text-[13px] text-[var(--xp-brand)]"
                          onClick={() => setActiveChip(section.type)}
                        >
                          See all
                        </button>
                      </div>
                      <div className="scrollbar-none flex gap-3 overflow-x-auto px-4 pb-1">
                        {section.items.map((item) => (
                          <div key={`${section.type}-${item.id}`} className="w-[160px] shrink-0">
                            <DiscoverItemCard
                              item={item}
                              userLat={location.lat}
                              userLng={location.lng}
                            />
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}

              {!feedLoading ? <div className="mx-4 h-px bg-[var(--xp-border)]" /> : null}
            </div>
          ) : null}

          <div className="mt-4 px-4">
            <h2 className="mb-3 text-[15px] font-semibold">{gridHeading}</h2>
            {itemsLoading ? (
              <GridSkeleton />
            ) : (
              <div className="grid grid-cols-2 gap-3">
                {items.map((item) => (
                  <DiscoverItemCard
                    key={item.id}
                    item={item}
                    userLat={location.lat}
                    userLng={location.lng}
                  />
                ))}
              </div>
            )}

            {!itemsLoading && items.length === 0 ? (
              <div className="xp-card px-6 py-10 text-center">
                <p className="text-lg font-semibold">Nothing matching that craving yet</p>
                <p className="mt-2 text-sm text-[var(--tg-theme-hint-color)]">
                  Try a different chip or check back soon.
                </p>
              </div>
            ) : null}
          </div>

          <div ref={sentinelRef} className="h-4" />

          {loadingMore ? (
            <div className="flex justify-center py-4">
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-[var(--xp-brand)] border-t-transparent" />
            </div>
          ) : null}

          {!hasMore && items.length > 0 ? (
            <p className="py-4 text-center text-[13px] text-[var(--tg-theme-hint-color)]">
              You&apos;ve seen everything
            </p>
          ) : null}
        </div>
      ) : null}

      {view === 'map' && selectedBranch ? (
        <BranchSheet branch={selectedBranch} onClose={() => setSelectedBranch(null)} />
      ) : null}

      <ViewToggle value={view} onChange={setView} />

      {cartCount > 0 ? (
        <CartBar
          count={cartCount}
          total={cartTotal}
          totalCartsCount={cart.totalCartsCount()}
          onOpen={() => navigate('/cart')}
        />
      ) : null}
    </div>
  )
}
