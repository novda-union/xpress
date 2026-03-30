export const TASHKENT_FALLBACK = {
  lat: 41.2995,
  lng: 69.2401,
}

export interface UserLocation {
  lat: number
  lng: number
  source: 'browser' | 'fallback'
}
