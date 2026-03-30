export function calculateDistanceInKm(
  fromLat: number,
  fromLng: number,
  toLat?: number | null,
  toLng?: number | null,
) {
  if (toLat == null || toLng == null) {
    return Number.POSITIVE_INFINITY
  }

  const earthRadiusKm = 6371
  const dLat = degreesToRadians(toLat - fromLat)
  const dLng = degreesToRadians(toLng - fromLng)
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(degreesToRadians(fromLat)) *
      Math.cos(degreesToRadians(toLat)) *
      Math.sin(dLng / 2) *
      Math.sin(dLng / 2)

  return earthRadiusKm * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

function degreesToRadians(value: number) {
  return value * (Math.PI / 180)
}

export function formatDistance(distanceKm: number) {
  if (!Number.isFinite(distanceKm)) {
    return 'Unknown'
  }
  if (distanceKm < 1) {
    return `${Math.round(distanceKm * 1000)} m`
  }
  return `${distanceKm.toFixed(1)} km`
}
