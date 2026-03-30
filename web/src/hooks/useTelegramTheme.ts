import { useEffect, useState } from 'react'
import { getTelegramColorScheme, initializeTelegram } from '../lib/telegram'

export function useTelegramTheme() {
  const [scheme] = useState<'light' | 'dark'>(() => {
    initializeTelegram()
    return getTelegramColorScheme()
  })

  useEffect(() => {
    document.documentElement.dataset.theme = scheme
  }, [scheme])

  return { scheme, isDark: scheme === 'dark' }
}
