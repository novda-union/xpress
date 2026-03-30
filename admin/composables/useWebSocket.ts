import type { AdminOrder } from '~/types/auth'

interface WSMessage {
  type: string
  order_id?: string
  order?: AdminOrder
  status?: string
  reason?: string
}

export function useAdminWebSocket() {
  const config = useRuntimeConfig()
  const { state: authState } = useAuth()
  const ws = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const callbacks = new Map<string, ((msg: WSMessage) => void)[]>()

  function connect() {
    if (!authState.token) return

    const wsUrl = config.public.apiBase.replace('http', 'ws')
    ws.value = new WebSocket(`${wsUrl}/admin/ws?token=${authState.token}`)

    ws.value.onopen = () => {
      isConnected.value = true
    }

    ws.value.onmessage = (event) => {
      const msg: WSMessage = JSON.parse(event.data)
      const handlers = callbacks.get(msg.type) || []
      handlers.forEach(cb => cb(msg))
    }

    ws.value.onclose = () => {
      isConnected.value = false
      // Reconnect after 3 seconds
      setTimeout(connect, 3000)
    }
  }

  function on(type: string, callback: (msg: WSMessage) => void) {
    if (!callbacks.has(type)) {
      callbacks.set(type, [])
    }
    callbacks.get(type)!.push(callback)
  }

  function disconnect() {
    ws.value?.close()
  }

  return { connect, disconnect, on, isConnected }
}
