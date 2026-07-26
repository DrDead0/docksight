import { useEffect, useRef, useState } from 'react'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { Container, LogEntry } from '@/types/api'

type ContainerLogsPanelProps = {
  hostId: string
  container: Container
  onClose: () => void
}

function resolveApiBase(): string {
  const raw = import.meta.env.VITE_API_URL as string | undefined
  if (!raw || raw.trim() === '') {
    return 'http://localhost:3000/api'
  }
  return raw.replace(/\/$/, '')
}

export function ContainerLogsPanel({
  hostId,
  container,
  onClose,
}: ContainerLogsPanelProps) {
  const [entries, setEntries] = useState<LogEntry[]>([])
  const [status, setStatus] = useState<'connecting' | 'live' | 'error'>(
    'connecting',
  )
  const [error, setError] = useState<string | null>(null)
  const scrollerRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const url = `${resolveApiBase()}/containers/${encodeURIComponent(container.id)}/logs?hostId=${encodeURIComponent(hostId)}&tail=100&follow=true`
    const source = new EventSource(url)

    const onChunk = (event: MessageEvent<string>) => {
      try {
        const payload = JSON.parse(event.data) as {
          entries?: LogEntry[]
        }
        const next = Array.isArray(payload.entries) ? payload.entries : []
        if (next.length === 0) {
          return
        }
        setEntries((current) => [...current, ...next].slice(-2_000))
        setStatus('live')
      } catch {
        // ignore malformed chunk
      }
    }

    source.addEventListener('logs.chunk', onChunk as EventListener)
    source.addEventListener('logs.subscribed', () => {
      setStatus('live')
    })
    source.onerror = () => {
      setStatus('error')
      setError('Log stream disconnected')
      source.close()
    }

    return () => {
      source.close()
    }
  }, [container.id, hostId])

  useEffect(() => {
    const node = scrollerRef.current
    if (node) {
      node.scrollTop = node.scrollHeight
    }
  }, [entries])

  return (
    <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-base font-medium tracking-tight">
            Logs · {container.name}
          </h3>
          <p className="text-xs text-muted-foreground">
            Live stream via agent · {status}
            {error ? ` · ${error}` : ''}
          </p>
        </div>
        <Button type="button" size="sm" variant="ghost" onClick={onClose}>
          <X className="h-4 w-4" aria-hidden />
          Close
        </Button>
      </div>

      <div
        ref={scrollerRef}
        className="h-80 overflow-auto rounded-md border border-border bg-background/80 p-3 font-mono text-xs leading-5"
      >
        {entries.length === 0 ? (
          <p className="text-muted-foreground">Waiting for log chunks…</p>
        ) : (
          entries.map((entry, index) => (
            <div key={`${entry.timestamp}-${index}`} className="whitespace-pre-wrap">
              <span className="text-muted-foreground">
                {entry.timestamp} [{entry.stream}]{' '}
              </span>
              <span
                className={
                  entry.stream === 'stderr' ? 'text-rose-300' : 'text-foreground'
                }
              >
                {entry.message}
              </span>
            </div>
          ))
        )}
      </div>
    </section>
  )
}
