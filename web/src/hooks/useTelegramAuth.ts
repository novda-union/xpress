import { useEffect, useState } from 'react'
import { api, setToken } from '../lib/api'
import { getInitDataRaw, initializeTelegram } from '../lib/telegram'
import type { AuthUser } from '../types'

interface AuthResponse {
  token: string
  user: AuthUser
}

export function useTelegramAuth() {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(Boolean(localStorage.getItem('xpressgo_token')))
  const [loading, setLoading] = useState(!localStorage.getItem('xpressgo_token'))

  useEffect(() => {
    initializeTelegram()
    if (isAuthenticated) return

    const initData = getInitDataRaw()
    const request = initData
      ? api<AuthResponse>('/auth/telegram', { method: 'POST', body: JSON.stringify({ init_data: initData }) })
      : api<AuthResponse>('/auth/dev', { method: 'POST', body: JSON.stringify({ telegram_id: 123456789 }) })

    request
      .then((response) => {
        setToken(response.token)
        setUser(response.user)
        setIsAuthenticated(true)
      })
      .finally(() => setLoading(false))
  }, [isAuthenticated])

  return { user, loading, isAuthenticated }
}
