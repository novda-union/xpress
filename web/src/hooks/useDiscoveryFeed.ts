import { useEffect, useRef, useState } from 'react'
import { api } from '../lib/api'
import type { FeedResponse, FeedSection } from '../types'

export function useDiscoveryFeed() {
  const [sections, setSections] = useState<FeedSection[]>([])
  const [loading, setLoading] = useState(true)
  const requestIdRef = useRef(0)

  useEffect(() => {
    const controller = new AbortController()
    const requestId = ++requestIdRef.current

    api<FeedResponse>('/discover/feed', { signal: controller.signal })
      .then((response) => {
        if (controller.signal.aborted || requestId !== requestIdRef.current) {
          return
        }
        setSections(response.sections)
      })
      .catch((error) => {
        if (controller.signal.aborted || requestId !== requestIdRef.current) {
          return
        }
        console.error('Failed to load discovery feed', error)
        setSections([])
      })
      .finally(() => {
        if (controller.signal.aborted || requestId !== requestIdRef.current) {
          return
        }
        setLoading(false)
      })

    return () => {
      controller.abort()
    }
  }, [])

  return { sections, loading }
}
