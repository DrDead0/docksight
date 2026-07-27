import { RefreshCw, X } from 'lucide-react'
import { StatusBadge } from '@/components/StatusBadge'
import { Button } from '@/components/ui/button'
import { useContainerInspect } from '@/hooks/useContainerInspect'
import { ApiError } from '@/services/api'
import type { Container } from '@/types/api'

type ContainerInspectPanelProps = {
  hostId: string
  container: Container
  onClose: () => void
}

function formatDate(value: string): string {
  if (!value) {
    return 'Unknown'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

function formatCommand(command: string[]): string {
  return command?.length > 0 ? command.join(' ') : 'None'
}

export function ContainerInspectPanel({
  hostId,
  container,
  onClose,
}: ContainerInspectPanelProps) {
  const inspectQuery = useContainerInspect(hostId, container.id)
  const result = inspectQuery.data
  const details = result?.container

  const errorMessage =
    inspectQuery.error instanceof ApiError
      ? inspectQuery.error.message
      : inspectQuery.error instanceof Error
        ? inspectQuery.error.message
        : result?.error

  return (
    <section className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="truncate text-base font-medium tracking-tight">
            Inspect {container.name}
          </h3>
          <p className="truncate text-xs text-muted-foreground">
            {details?.image ?? container.image}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={() => void inspectQuery.refetch()}
            disabled={inspectQuery.isFetching}
          >
            <RefreshCw
              className={
                inspectQuery.isFetching
                  ? 'h-4 w-4 animate-spin'
                  : 'h-4 w-4'
              }
              aria-hidden
            />
            Refresh
          </Button>
          <Button type="button" size="sm" variant="ghost" onClick={onClose}>
            <X className="h-4 w-4" aria-hidden />
            Close
          </Button>
        </div>
      </div>

      {inspectQuery.isLoading ? (
        <p className="text-sm text-muted-foreground">Loading inspect data...</p>
      ) : errorMessage || !details ? (
        <div className="rounded-lg border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
          {errorMessage ?? 'Container inspect failed'}
        </div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-[1fr_1fr]">
          <div className="space-y-4">
            <InfoGrid
              items={[
                ['Container ID', details.shortId || details.id],
                ['Name', details.name],
                ['Created', formatDate(details.created)],
                ['Started', formatDate(details.startedAt)],
                ['Working dir', details.workingDir || 'None'],
                ['Restart policy', details.restartPolicy || 'None'],
              ]}
            />

            <section className="rounded-md border border-border bg-background/60 p-3">
              <div className="mb-3 flex items-center justify-between gap-3">
                <h4 className="text-sm font-medium">State</h4>
                <StatusBadge status={details.state.status} />
              </div>
              <InfoGrid
                compact
                items={[
                  ['Running', details.state.running ? 'Yes' : 'No'],
                  ['Paused', details.state.paused ? 'Yes' : 'No'],
                  ['Restarting', details.state.restarting ? 'Yes' : 'No'],
                ]}
              />
            </section>

            <section className="rounded-md border border-border bg-background/60 p-3">
              <h4 className="mb-2 text-sm font-medium">Command</h4>
              <p className="break-words font-mono text-xs text-muted-foreground">
                {formatCommand(details.cmd)}
              </p>
            </section>
          </div>

          <div className="space-y-4">
            <ListSection
              title="Ports"
              empty="No published ports"
              rows={details.ports.map((port) => ({
                key: `${port.private}-${port.public}-${port.protocol}`,
                primary: `${port.private}/${port.protocol}`,
                secondary: port.public ? `Published ${port.public}` : 'Internal only',
              }))}
            />
            <ListSection
              title="Mounts"
              empty="No mounts"
              rows={details.mounts.map((mount) => ({
                key: `${mount.source}-${mount.target}`,
                primary: mount.target,
                secondary: `${mount.source || 'Anonymous'} ${mount.mode ? `(${mount.mode})` : ''}`,
              }))}
            />
            <ListSection
              title="Networks"
              empty="No networks"
              rows={details.networks.map((network) => ({
                key: network.name,
                primary: network.name,
                secondary: network.ip || 'No IP assigned',
              }))}
            />
          </div>
        </div>
      )}
    </section>
  )
}

function InfoGrid({
  items,
  compact = false,
}: {
  items: Array<[string, string]>
  compact?: boolean
}) {
  return (
    <dl
      className={
        compact
          ? 'grid grid-cols-3 gap-3 text-xs'
          : 'grid grid-cols-1 gap-3 rounded-md border border-border bg-background/60 p-3 text-xs sm:grid-cols-2'
      }
    >
      {items.map(([label, value]) => (
        <div key={label} className="min-w-0">
          <dt className="text-muted-foreground">{label}</dt>
          <dd className="mt-0.5 truncate font-medium">{value || 'Unknown'}</dd>
        </div>
      ))}
    </dl>
  )
}

function ListSection({
  title,
  empty,
  rows,
}: {
  title: string
  empty: string
  rows: Array<{ key: string; primary: string; secondary: string }>
}) {
  return (
    <section className="rounded-md border border-border bg-background/60 p-3">
      <h4 className="mb-2 text-sm font-medium">{title}</h4>
      {rows.length === 0 ? (
        <p className="text-xs text-muted-foreground">{empty}</p>
      ) : (
        <div className="space-y-2">
          {rows.map((row) => (
            <div key={row.key} className="min-w-0 text-xs">
              <p className="truncate font-medium">{row.primary}</p>
              <p className="truncate text-muted-foreground">{row.secondary}</p>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
