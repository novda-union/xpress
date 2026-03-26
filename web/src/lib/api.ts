const API_BASE = 'http://localhost:8080'

function getToken(): string | null {
  return localStorage.getItem('xpressgo_token')
}

export function setToken(token: string) {
  localStorage.setItem('xpressgo_token', token)
}

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || 'Request failed')
  }

  return res.json()
}

export function getWsUrl(): string {
  const token = getToken()
  return `ws://localhost:8080/ws?token=${token}`
}
