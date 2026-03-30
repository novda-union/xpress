import { lazy, Suspense } from 'react'
import { UtensilsCrossed } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { PhoneGate } from '../components/auth/PhoneGate'
import { AppShell } from '../components/layout/AppShell'
import { BranchListCard } from '../components/discovery/BranchListCard'
import { BranchSheet } from '../components/discovery/BranchSheet'
import { CategoryTabs } from '../components/discovery/CategoryTabs'
import { ViewToggle } from '../components/discovery/ViewToggle'
import { useDiscovery } from '../hooks/useDiscovery'
import { useTelegramAuth } from '../hooks/useTelegramAuth'
import { useTelegramTheme } from '../hooks/useTelegramTheme'
import { useUserLocation } from '../hooks/useUserLocation'
import type { DiscoverBranch } from '../types'

const DiscoveryMap = lazy(() =>
  import('../components/discovery/DiscoveryMap').then((module) => ({ default: module.DiscoveryMap })),
)

export default function HomePage() {
  useTelegramTheme()

  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const storeSlug = searchParams.get('store')
  const auth = useTelegramAuth()
  const location = useUserLocation()
  const discovery = useDiscovery(location.lat, location.lng)
  const [view, setView] = useState<'map' | 'list'>(storeSlug ? 'list' : 'map')
  const [selectedBranch, setSelectedBranch] = useState<(DiscoverBranch & { distanceKm: number }) | null>(null)

  const visibleBranches = useMemo(
    () => discovery.branches.filter((branch) => !storeSlug || branch.store_slug === storeSlug),
    [discovery.branches, storeSlug],
  )

  const featuredStore = visibleBranches[0]?.store_name

  const topBar = useMemo(
    () => (
      <div className="xp-page-padding pt-4">
        <div className="glass-panel flex items-center justify-between rounded-[24px] px-4 py-3 shadow-sm">
          <div className="flex items-center gap-2">
            <span className="flex h-10 w-10 items-center justify-center rounded-2xl bg-[var(--xp-brand-muted)] text-[var(--xp-brand)]">
              <UtensilsCrossed className="h-5 w-5" />
            </span>
            <div>
              <p className="text-sm font-semibold">Xpressgo</p>
              <p className="text-xs text-[var(--tg-theme-hint-color)]">
                {storeSlug && featuredStore
                  ? `${featuredStore} branches`
                  : location.source === 'browser'
                    ? 'Nearby branches'
                    : 'Showing Tashkent nearby'}
              </p>
            </div>
          </div>
          <ViewToggle value={view} onChange={setView} />
        </div>
      </div>
    ),
    [featuredStore, location.source, storeSlug, view],
  )

  if (!auth.isAuthenticated) {
    return <PhoneGate loading={auth.loading} error={auth.error} onRequestAccess={auth.requestAccess} />
  }

  return (
    <AppShell header={topBar}>
      <CategoryTabs value={discovery.category} onChange={discovery.setCategory} />
      <div className="xp-page-padding pb-6">
        {view === 'map' ? (
          <Suspense fallback={<div className="h-[calc(100vh-5rem)] w-full animate-pulse rounded-[24px] bg-[var(--xp-card-bg)]" />}>
            <DiscoveryMap
              branches={visibleBranches}
              center={location}
              selectedBranchId={selectedBranch?.branch_id ?? null}
              visible
              onSelect={setSelectedBranch}
            />
          </Suspense>
        ) : null}

        <div className={`${view === 'list' ? 'block' : 'hidden'} space-y-4`}>
          {visibleBranches.map((branch) => (
            <BranchListCard
              key={branch.branch_id}
              branch={branch}
              onSelect={(selected) => navigate(`/branch/${selected.branch_id}`)}
            />
          ))}
          {visibleBranches.length === 0 ? (
            <div className="xp-card px-6 py-10 text-center">
              <p className="text-lg font-semibold">No matching branches</p>
              <p className="mt-2 text-sm text-[var(--tg-theme-hint-color)]">
                Try another store link or return to discovery.
              </p>
            </div>
          ) : null}
        </div>
      </div>

      <BranchSheet branch={selectedBranch} onClose={() => setSelectedBranch(null)} />
    </AppShell>
  )
}
