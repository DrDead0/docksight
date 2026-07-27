import { useCallback, useEffect, useRef, useState } from 'react'
import type { LogEntry } from '@/types/api'

export type LogStreamStatus = 'connecting' | 'live' | 'paused' | 'error'

const MAX_ENTRIES = 5_000

function resolveApiBase(): string {
  const raw = import.meta.env.VITE_API_URL as string | undefined
  if (!raw || raw.trim() === '') {
    return 'http://localhost:3000/api'
  }
  return raw.replace(/\/$/, '')
}

/**
 * Subscribes to `GET /containers/:id/logs` (SSE). Pausing keeps the stream
 * open and buffers entries so nothing is lost while the user reads.
 */
export function useContainerLogs(
  hostId: string | undefined,
  containerId: string | undefined,
  { tail = 200 }: { tail?: number } = {},
) {
  const [entries, setEntries] = useState<LogEntry[]>([])
  const [status, setStatus] = useState<LogStreamStatus>('connecting')
  const [error, setError] = useState<string | null>(null)
  const [paused, setPaused] = useState(false)
  const [bufferedCount, setBufferedCount] = useState(0)

  const pausedRef = useRef(paused)
  const bufferRef = useRef<LogEntry[]>([])
  const [reconnectToken, setReconnectToken] = useState(0)

  useEffect(() => {
    pausedRef.current = paused
  }, [paused])

  useEffect(() => {
    if (!hostId || !containerId) {
      return
    }

    setEntries([])
    setError(null)
    setStatus('connecting')
    bufferRef.current = []
    setBufferedCount(0)

    const url = `${resolveApiBase()}/containers/${encodeURIComponent(containerId)}/logs?hostId=${encodeURIComponent(hostId)}&tail=${tail}&follow=true`
    const source = new EventSource(url)

    const onChunk = (event: MessageEvent<string>) => {
      let next: LogEntry[] = []
      try {
        const payload = JSON.parse(event.data) as { entries?: LogEntry[] }
        next = Array.isArray(payload.entries) ? payload.entries : []
      } catch {
        return
      }
      if (next.length === 0) {
        return
      }

      if (pausedRef.current) {
        bufferRef.current = [...bufferRef.current, ...next].slice(-MAX_ENTRIES)
        setBufferedCount(bufferRef.current.length)
        return
      }

      setEntries((current) => [...current, ...next].slice(-MAX_ENTRIES))
      setStatus('live')
    }

    source.addEventListener('logs.chunk', onChunk as EventListener)
    source.addEventListener('logs.subscribed', () => {
      setStatus((current) => (current === 'paused' ? current : 'live'))
    })
    source.onerror = () => {
      setStatus('error')
      setError('Log stream disconnected')
      source.close()
    }

    return () => {
      source.close()
    }
  }, [hostId, containerId, tail, reconnectToken])

  const togglePause = useCallback(() => {
    setPaused((current) => {
      const next = !current
      if (!next && bufferRef.current.length > 0) {
        const buffered = bufferRef.current
        bufferRef.current = []
        setBufferedCount(0)
        setEntries((entriesNow) =>
          [...entriesNow, ...buffered].slice(-MAX_ENTRIES),
        )
      }
      setStatus(next ? 'paused' : 'live')
      return next
    })
  }, [])

  const clear = useCallback(() => {
    setEntries([])
    bufferRef.current = []
    setBufferedCount(0)
  }, [])

  const reconnect = useCallback(() => {
    setReconnectToken((token) => token + 1)
  }, [])

  return {
    entries,
    status,
    error,
    paused,
    bufferedCount,
    togglePause,
    clear,
    reconnect,
  }
}
