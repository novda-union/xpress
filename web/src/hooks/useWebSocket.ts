import { useEffect, useRef, useCallback, useState } from 'react'
import { getWsUrl } from '../lib/api'

interface WSMessage {
  type: string
  order_id?: string
  status?: string
  reason?: string
  order?: any
}

export function useWebSocket(onMessage: (msg: WSMessage) => void) {
  const wsRef = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)

  const connect = useCallback(() => {
    const ws = new WebSocket(getWsUrl())
    wsRef.current = ws

    ws.onopen = () => {
      setIsConnected(true)
    }

    ws.onmessage = (event) => {
      const msg: WSMessage = JSON.parse(event.data)
      onMessage(msg)
    }

    ws.onclose = () => {
      setIsConnected(false)
      setTimeout(connect, 3000)
    }

    return ws
  }, [onMessage])

  useEffect(() => {
    const ws = connect()
    return () => ws.close()
  }, [connect])

  return { isConnected }
}
