import { useEffect, useState } from 'react'
import { api, setToken } from '../lib/api'
import { getInitDataRaw, initializeTelegram, requestTelegramContact } from '../lib/telegram'
import type { AuthUser } from '../types'

interface AuthResponse {
  token: string
  user: AuthUser
}

export function useTelegramAuth() {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [isAuthenticated, setIsAuthenticated] = useState(Boolean(localStorage.getItem('xpressgo_token')))

  useEffect(() => {
    initializeTelegram()

    if (localStorage.getItem('xpressgo_token')) {
      setLoading(false)
      setIsAuthenticated(true)
      return
    }

    autoAuthenticate()
      .catch(() => {
        setLoading(false)
      })
  }, [])

  async function autoAuthenticate() {
    const initData = getInitDataRaw()
    if (initData) {
      const response = await api<AuthResponse>('/auth/telegram', {
        method: 'POST',
        body: JSON.stringify({ init_data: initData }),
      })
      setToken(response.token)
      setUser(response.user)
      setIsAuthenticated(true)
      setError('')
      setLoading(false)
      return
    }

    try {
      const response = await api<AuthResponse>('/auth/dev', {
        method: 'POST',
        body: JSON.stringify({ telegram_id: 123456789 }),
      })
      setToken(response.token)
      setUser(response.user)
      setIsAuthenticated(true)
    } finally {
      setLoading(false)
    }
  }

  async function requestAccess() {
    setLoading(true)
    setError('')
    try {
      const phone = await requestTelegramContact()
      const initData = getInitDataRaw()

      let response: AuthResponse
      if (initData) {
        response = await api<AuthResponse>('/auth/telegram', {
          method: 'POST',
          body: JSON.stringify({
            init_data: initData,
            phone_number: phone,
          }),
        })
      } else {
        response = await api<AuthResponse>('/auth/dev', {
          method: 'POST',
          body: JSON.stringify({ telegram_id: 123456789 }),
        })
      }

      setToken(response.token)
      setUser(response.user)
      setIsAuthenticated(true)
      return true
    } catch (requestError) {
      const message = requestError instanceof Error ? requestError.message : 'Phone number is required to continue'
      setError(message)
      setIsAuthenticated(false)
      return false
    } finally {
      setLoading(false)
    }
  }

  return {
    user,
    error,
    loading,
    isAuthenticated,
    requestAccess,
  }
}
