import { useEffect, useMemo, useState } from 'react'
import { calculateDistanceInKm } from '../lib/distance'
import { api } from '../lib/api'
import type { DiscoverBranch, StoreCategory } from '../types'

export type DiscoveryCategory = 'all' | StoreCategory

export function useDiscovery(lat: number, lng: number) {
  const [branches, setBranches] = useState<DiscoverBranch[]>([])
  const [category, setCategory] = useState<DiscoveryCategory>('all')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let active = true

    api<DiscoverBranch[]>('/discover')
      .then((nextBranches) => {
        if (active) {
          setBranches(nextBranches)
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })

    return () => {
      active = false
    }
  }, [])

  const filteredBranches = useMemo(() => {
    return branches
      .filter((branch) => category === 'all' || branch.store_category === category)
      .map((branch) => ({
        ...branch,
        distanceKm: calculateDistanceInKm(lat, lng, branch.lat, branch.lng),
      }))
      .sort((left, right) => left.distanceKm - right.distanceKm)
  }, [branches, category, lat, lng])

  return {
    branches: filteredBranches,
    category,
    loading,
    setCategory,
  }
}
