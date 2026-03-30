import {
  bindThemeParamsCssVars,
  bindViewportCssVars,
  init,
  miniAppReady,
  mountMiniAppSync,
  mountThemeParamsSync,
  mountViewport,
  requestContact,
  retrieveRawInitData,
} from '@telegram-apps/sdk'

declare global {
  interface Window {
    Telegram?: {
      WebApp?: {
        ready: () => void
        expand: () => void
        initData?: string
        colorScheme?: 'light' | 'dark'
      }
    }
  }
}

let initialized = false

export function initializeTelegram() {
  if (initialized || typeof window === 'undefined') {
    return
  }

  try {
    init()
    mountMiniAppSync()
    mountThemeParamsSync()
    mountViewport()
    bindThemeParamsCssVars()
    bindViewportCssVars()
    miniAppReady()
  } catch {
    // Browser preview outside Telegram is allowed.
  }

  window.Telegram?.WebApp?.ready?.()
  window.Telegram?.WebApp?.expand?.()
  initialized = true
}

export function getInitDataRaw() {
  try {
    const raw = retrieveRawInitData()
    if (raw) {
      return raw
    }
  } catch {
    // Fall back to WebApp initData below.
  }

  return window.Telegram?.WebApp?.initData ?? ''
}

export function getTelegramColorScheme() {
  return window.Telegram?.WebApp?.colorScheme ?? 'light'
}

export async function requestTelegramContact() {
  if (!requestContact.isAvailable()) {
    throw new Error('Phone sharing is unavailable in this environment')
  }
  const result = await requestContact()
  return result.contact.phone_number
}
