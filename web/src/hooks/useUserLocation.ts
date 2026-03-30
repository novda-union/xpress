import { useEffect, useState } from 'react'
import { TASHKENT_FALLBACK, type UserLocation } from '../lib/location'

export function useUserLocation() {
  const [location, setLocation] = useState<UserLocation>({
    ...TASHKENT_FALLBACK,
    source: 'fallback',
  })

  useEffect(() => {
    if (!navigator.geolocation) {
      return
    }

    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        setLocation({
          lat: coords.latitude,
          lng: coords.longitude,
          source: 'browser',
        })
      },
      () => {
        setLocation({
          ...TASHKENT_FALLBACK,
          source: 'fallback',
        })
      },
      {
        enableHighAccuracy: true,
        timeout: 5000,
      },
    )
  }, [])

  return location
}
