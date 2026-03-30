const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

function getToken(): string | null {
  return localStorage.getItem('xpressgo_token')
}

export function setToken(token: string) {
  localStorage.setItem('xpressgo_token', token)
}

export function clearToken() {
  localStorage.removeItem('xpressgo_token')
}

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers = new Headers(options.headers)

  if (!headers.has('Content-Type') && options.body && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || 'Request failed')
  }

  if (res.status === 204) {
    return undefined as T
  }

  return res.json()
}

export function getWsUrl(): string {
  const token = getToken()
  const base = API_BASE.replace(/^http/, 'ws')
  return `${base}/ws?token=${token ?? ''}`
}
