import { cn } from '@/lib/utils'

type StatusBadgeProps = {
  status: string
  className?: string
}

function normalizeStatus(status: string): string {
  return status.trim().toUpperCase()
}

function badgeClasses(status: string): string {
  const value = normalizeStatus(status)

  if (
    value === 'ONLINE' ||
    value === 'RUNNING' ||
    value.startsWith('UP ') ||
    value === 'UP'
  ) {
    return 'border-emerald-500/30 bg-emerald-500/15 text-emerald-300'
  }

  if (
    value === 'OFFLINE' ||
    value === 'EXITED' ||
    value === 'DEAD' ||
    value.startsWith('EXITED')
  ) {
    return 'border-rose-500/30 bg-rose-500/15 text-rose-300'
  }

  if (value === 'RESTARTING' || value === 'PAUSED' || value === 'CREATED') {
    return 'border-amber-500/30 bg-amber-500/15 text-amber-300'
  }

  return 'border-border bg-muted text-muted-foreground'
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium tracking-wide',
        badgeClasses(status),
        className,
      )}
    >
      {status}
    </span>
  )
}
