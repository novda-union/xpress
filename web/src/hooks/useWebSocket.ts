import { useEffect, useRef, useState } from 'react'
import { getWsUrl } from '../lib/api'
import type { Order } from '../types'

interface WSMessage {
  type: string
  order_id?: string
  status?: string
  reason?: string
  order?: Order
}

export function useWebSocket(onMessage: (msg: WSMessage) => void) {
  const reconnectTimeoutRef = useRef<number | null>(null)
  const onMessageRef = useRef(onMessage)
  const [isConnected, setIsConnected] = useState(false)

  useEffect(() => {
    onMessageRef.current = onMessage
  }, [onMessage])

  useEffect(() => {
    let cancelled = false

    const connect = () => {
      const ws = new WebSocket(getWsUrl())

      ws.onopen = () => {
        setIsConnected(true)
      }

      ws.onmessage = (event) => {
        const msg: WSMessage = JSON.parse(event.data)
        onMessageRef.current(msg)
      }

      ws.onclose = () => {
        setIsConnected(false)
        if (cancelled) {
          return
        }
        reconnectTimeoutRef.current = window.setTimeout(() => {
          connect()
        }, 3000)
      }

      return ws
    }

    const ws = connect()
    return () => {
      cancelled = true
      if (reconnectTimeoutRef.current !== null) {
        window.clearTimeout(reconnectTimeoutRef.current)
      }
      ws.close()
    }
  }, [])

  return { isConnected }
}
