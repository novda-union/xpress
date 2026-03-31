import { useCallback, useEffect, useRef, useState } from 'react'
import { api } from '../lib/api'
import type { DiscoverItem, ItemsPageResponse } from '../types'

const PAGE_SIZE = 20

export function useDiscoveryItems(category: string, sort: string) {
  const queryKey = `${category}:${sort}`
  const [items, setItems] = useState<DiscoverItem[]>([])
  const [settledQueryKey, setSettledQueryKey] = useState(queryKey)
  const [loadingMore, setLoadingMore] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [total, setTotal] = useState(0)

  const nextPageRef = useRef(1)
  const hasMoreRef = useRef(true)
  const loadingMoreRef = useRef(false)
  const requestIdRef = useRef(0)

  const syncHasMore = (nextValue: boolean) => {
    hasMoreRef.current = nextValue
    setHasMore(nextValue)
  }

  const syncLoadingMore = (nextValue: boolean) => {
    loadingMoreRef.current = nextValue
    setLoadingMore(nextValue)
  }

  useEffect(() => {
    const controller = new AbortController()
    const requestId = ++requestIdRef.current

    nextPageRef.current = 1
    hasMoreRef.current = true
    loadingMoreRef.current = false

    const params = new URLSearchParams({
      page: '1',
      limit: String(PAGE_SIZE),
      sort,
    })
    if (category) {
      params.set('category', category)
    }

    api<ItemsPageResponse>(`/discover/items?${params.toString()}`, { signal: controller.signal })
      .then((response) => {
        if (controller.signal.aborted || requestId !== requestIdRef.current) {
          return
        }

        setSettledQueryKey(queryKey)
        setItems(response.items)
        setTotal(response.total)
        syncHasMore(response.page * response.limit < response.total)
        nextPageRef.current = response.page + 1
      })
      .catch((error) => {
        if (controller.signal.aborted || requestId !== requestIdRef.current) {
          return
        }
        console.error('Failed to load discovery items', error)
        setSettledQueryKey(queryKey)
        setItems([])
        setTotal(0)
        syncHasMore(false)
      })

    return () => {
      controller.abort()
    }
  }, [category, queryKey, sort])

  const loadMore = useCallback(() => {
    if (settledQueryKey !== queryKey || loadingMoreRef.current || !hasMoreRef.current) {
      return
    }

    const controller = new AbortController()
    const requestId = ++requestIdRef.current
    syncLoadingMore(true)

    const params = new URLSearchParams({
      page: String(nextPageRef.current),
      limit: String(PAGE_SIZE),
      sort,
    })
    if (category) {
      params.set('category', category)
    }

    api<ItemsPageResponse>(`/discover/items?${params.toString()}`, { signal: controller.signal })
      .then((response) => {
        if (controller.signal.aborted || requestId !== requestIdRef.current) {
          return
        }

        setItems((current) => {
          const merged = [...current, ...response.items]
          return merged
        })
        setTotal(response.total)
        syncHasMore(response.page * response.limit < response.total)
        nextPageRef.current = response.page + 1
      })
      .catch((error) => {
        if (controller.signal.aborted || requestId !== requestIdRef.current) {
          return
        }
        console.error('Failed to load more discovery items', error)
      })
      .finally(() => {
        if (controller.signal.aborted || requestId !== requestIdRef.current) {
          return
        }
        syncLoadingMore(false)
      })
  }, [category, queryKey, settledQueryKey, sort])

  const loading = settledQueryKey !== queryKey

  return {
    items: loading ? [] : items,
    loading,
    loadingMore,
    hasMore: loading ? true : hasMore,
    total: loading ? 0 : total,
    loadMore,
  }
}
