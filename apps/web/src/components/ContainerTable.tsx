import type { ReactNode } from 'react'
import { LoaderCircle, Play, RotateCcw, Square } from 'lucide-react'
import { StatusBadge } from '@/components/StatusBadge'
import { Button } from '@/components/ui/button'
import type { Container, ContainerAction } from '@/types/api'

type ContainerTableProps = {
  containers: Container[]
  busyKey?: string | null
  onAction?: (container: Container, action: ContainerAction) => void
}

function isRunning(container: Container): boolean {
  const state = (container.state || '').toLowerCase()
  return state === 'running' || state === 'restarting'
}

export function ContainerTable({
  containers,
  busyKey = null,
  onAction,
}: ContainerTableProps) {
  if (containers.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-border px-4 py-10 text-center text-sm text-muted-foreground">
        No containers discovered for this host yet.
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-border">
      <table className="w-full min-w-[44rem] border-collapse text-left text-sm">
        <thead className="bg-muted/40 text-xs uppercase tracking-wide text-muted-foreground">
          <tr>
            <th className="px-4 py-3 font-medium">Name</th>
            <th className="px-4 py-3 font-medium">Image</th>
            <th className="px-4 py-3 font-medium">Status</th>
            <th className="px-4 py-3 font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {containers.map((container) => {
            const running = isRunning(container)
            const rowBusy =
              busyKey?.startsWith(`${container.id}:`) ?? false
            const busyAction = busyKey?.split(':')[1] as
              | ContainerAction
              | undefined

            return (
              <tr
                key={container.id}
                className="border-t border-border/80 hover:bg-accent/30"
              >
                <td className="px-4 py-3 font-medium">{container.name}</td>
                <td className="px-4 py-3 text-muted-foreground">
                  {container.image}
                </td>
                <td className="px-4 py-3">
                  <StatusBadge status={container.state || container.status} />
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <ActionButton
                      label="Start"
                      icon={<Play className="h-3.5 w-3.5" aria-hidden />}
                      disabled={running || rowBusy || !onAction}
                      loading={rowBusy && busyAction === 'start'}
                      onClick={() => onAction?.(container, 'start')}
                    />
                    <ActionButton
                      label="Stop"
                      icon={<Square className="h-3.5 w-3.5" aria-hidden />}
                      disabled={!running || rowBusy || !onAction}
                      loading={rowBusy && busyAction === 'stop'}
                      onClick={() => onAction?.(container, 'stop')}
                    />
                    <ActionButton
                      label="Restart"
                      icon={<RotateCcw className="h-3.5 w-3.5" aria-hidden />}
                      disabled={!running || rowBusy || !onAction}
                      loading={rowBusy && busyAction === 'restart'}
                      onClick={() => onAction?.(container, 'restart')}
                    />
                  </div>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

function ActionButton({
  label,
  icon,
  disabled,
  loading,
  onClick,
}: {
  label: string
  icon: ReactNode
  disabled?: boolean
  loading?: boolean
  onClick: () => void
}) {
  return (
    <Button
      type="button"
      size="sm"
      variant="outline"
      disabled={disabled}
      onClick={onClick}
      aria-label={label}
    >
      {loading ? (
        <LoaderCircle className="h-3.5 w-3.5 animate-spin" aria-hidden />
      ) : (
        icon
      )}
      {label}
    </Button>
  )
}
