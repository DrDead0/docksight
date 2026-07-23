import { Server } from 'lucide-react'
import { StatusBadge } from '@/components/StatusBadge'
import type { Host } from '@/types/api'
import { cn } from '@/lib/utils'

type HostCardProps = {
  host: Host
  selected?: boolean
  onSelect?: (hostId: string) => void
}

function formatLastSeen(lastSeen: string | null): string {
  if (!lastSeen) {
    return 'Never'
  }
  const date = new Date(lastSeen)
  if (Number.isNaN(date.getTime())) {
    return lastSeen
  }
  return date.toLocaleString()
}

export function HostCard({ host, selected = false, onSelect }: HostCardProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect?.(host.id)}
      className={cn(
        'w-full rounded-lg border bg-card p-4 text-left transition-colors',
        'hover:border-primary/40 hover:bg-accent/40',
        selected ? 'border-primary/60 ring-1 ring-primary/40' : 'border-border',
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <div className="rounded-md bg-secondary p-2 text-primary">
            <Server className="h-4 w-4" aria-hidden />
          </div>
          <div className="min-w-0">
            <p className="truncate font-medium tracking-tight">{host.hostname}</p>
            <p className="truncate text-xs text-muted-foreground">
              {host.os} · {host.architecture}
            </p>
          </div>
        </div>
        <StatusBadge status={host.status} />
      </div>

      <dl className="mt-4 grid grid-cols-2 gap-3 text-xs">
        <div>
          <dt className="text-muted-foreground">Version</dt>
          <dd className="mt-0.5 font-medium">{host.version}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Last seen</dt>
          <dd className="mt-0.5 font-medium">{formatLastSeen(host.lastSeen)}</dd>
        </div>
      </dl>
    </button>
  )
}
