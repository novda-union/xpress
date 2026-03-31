import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'
import type { DiscoverBranch } from '../../types'

interface DiscoveryMapProps {
  branches: Array<DiscoverBranch & { distanceKm: number }>
  selectedBranchId?: string | null
  center: { lat: number; lng: number }
  visible: boolean
  onSelect: (branch: DiscoverBranch & { distanceKm: number }) => void
}

export function DiscoveryMap({ branches, center, selectedBranchId, visible, onSelect }: DiscoveryMapProps) {
  const mapRef = useRef<maplibregl.Map | null>(null)
  const containerRef = useRef<HTMLDivElement | null>(null)
  const markersRef = useRef<maplibregl.Marker[]>([])

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return
    }

    mapRef.current = new maplibregl.Map({
      container: containerRef.current,
      style: 'https://tiles.openfreemap.org/styles/liberty',
      center: [center.lng, center.lat],
      zoom: 11.3,
      attributionControl: false,
    })

    mapRef.current.addControl(new maplibregl.NavigationControl({ showCompass: false }), 'top-right')

    return () => {
      markersRef.current.forEach((marker) => marker.remove())
      mapRef.current?.remove()
      mapRef.current = null
    }
  }, [center.lat, center.lng])

  useEffect(() => {
    if (!mapRef.current) {
      return
    }
    mapRef.current.setCenter([center.lng, center.lat])
  }, [center.lat, center.lng])

  useEffect(() => {
    markersRef.current.forEach((marker) => marker.remove())
    markersRef.current = []

    if (!mapRef.current) {
      return
    }

    branches.forEach((branch) => {
      if (branch.lat == null || branch.lng == null) {
        return
      }

      const element = document.createElement('button')
      element.type = 'button'
      element.className = `marker-pin ${branch.branch_id === selectedBranchId ? 'is-selected' : ''}`
      element.innerHTML = `<img src="${branch.store_logo_url || 'https://placehold.co/88x88?text=XG'}" alt="${branch.store_name}" />`
      element.addEventListener('click', () => onSelect(branch))

      const marker = new maplibregl.Marker({ element })
        .setLngLat([branch.lng, branch.lat])
        .addTo(mapRef.current!)

      markersRef.current.push(marker)
    })
  }, [branches, onSelect, selectedBranchId])

  return (
    <div className={`${visible ? 'block' : 'hidden'} h-[calc(100vh-5rem)] w-full overflow-hidden rounded-[24px]`}>
      <div ref={containerRef} className="h-full w-full" />
    </div>
  )
}
