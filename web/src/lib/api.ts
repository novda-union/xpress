const API_BASE =
  import.meta.env.VITE_API_BASE_URL?.trim() ?? (import.meta.env.PROD ? 'https://srvr.novdaunion.uz' : '')

function getToken(): string | null {
  return localStorage.getItem('xpressgo_token')
}

export function setToken(token: string) {
  localStorage.setItem('xpressgo_token', token)
}

export function clearToken() {
  localStorage.removeItem('xpressgo_token')
}

function resolveApiUrl(path: string): string {
  if (!API_BASE) {
    return path
  }

  return new URL(path, API_BASE).toString()
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

  const res = await fetch(resolveApiUrl(path), {
    ...options,
    headers,
  })

  if (!res.ok) {
    if (res.status === 401) {
      localStorage.removeItem('xpressgo_token')
    }
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
  const url = API_BASE ? new URL('/ws', API_BASE) : new URL('/ws', window.location.origin)

  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.searchParams.set('token', token ?? '')

  return url.toString()
}
