import { useEffect, useMemo, useState } from 'react'
import { Container, RefreshCw } from 'lucide-react'
import { ContainerLogsPanel } from '@/components/ContainerLogsPanel'
import { ContainerTable } from '@/components/ContainerTable'
import { HostCard } from '@/components/HostCard'
import { useToast } from '@/components/ToastProvider'
import { Button } from '@/components/ui/button'
import { useContainerAction } from '@/hooks/useContainerAction'
import { useContainers } from '@/hooks/useContainers'
import { useHosts } from '@/hooks/useHosts'
import { ApiError } from '@/services/api'
import type { Container as ContainerRow, ContainerAction } from '@/types/api'

export function DashboardPage() {
  const toast = useToast()
  const hostsQuery = useHosts()
  const hosts = hostsQuery.data ?? []
  const [selectedHostId, setSelectedHostId] = useState<string | undefined>()
  const [busyKey, setBusyKey] = useState<string | null>(null)
  const [logsContainer, setLogsContainer] = useState<ContainerRow | null>(null)

  useEffect(() => {
    if (hosts.length === 0) {
      setSelectedHostId(undefined)
      return
    }
    if (!selectedHostId || !hosts.some((host) => host.id === selectedHostId)) {
      setSelectedHostId(hosts[0].id)
      setLogsContainer(null)
    }
  }, [hosts, selectedHostId])

  const containersQuery = useContainers(selectedHostId)
  const containerAction = useContainerAction()
  const selectedHost = useMemo(
    () => hosts.find((host) => host.id === selectedHostId),
    [hosts, selectedHostId],
  )

  const onlineCount = hosts.filter((host) => host.status === 'ONLINE').length

  async function handleContainerAction(
    container: ContainerRow,
    action: ContainerAction,
  ) {
    if (!selectedHostId) {
      return
    }

    setBusyKey(`${container.id}:${action}`)
    try {
      const result = await containerAction.mutateAsync({
        containerId: container.id,
        hostId: selectedHostId,
        action,
        containerName: container.name,
      })

      if (result.ok) {
        toast.push({
          tone: 'success',
          title: `${action} succeeded`,
          description: `${container.name}: ${result.message}`,
        })
        await containersQuery.refetch()
      } else {
        toast.push({
          tone: 'error',
          title: `${action} failed`,
          description:
            result.error ?? result.message ?? `Could not ${action} container`,
        })
      }
    } catch (error) {
      const message =
        error instanceof ApiError
          ? error.message
          : error instanceof Error
            ? error.message
            : `Failed to ${action} container`
      toast.push({
        tone: 'error',
        title: `${action} failed`,
        description: message,
      })
    } finally {
      setBusyKey(null)
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-8 px-6 py-8">
      <section className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            Dashboard
          </h1>
          <p className="max-w-2xl text-sm text-muted-foreground">
            Connected Docker hosts and containers. Start, stop, restart, or open
            live logs from the table.
          </p>
        </div>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          onClick={() => {
            void hostsQuery.refetch()
            void containersQuery.refetch()
          }}
          disabled={hostsQuery.isFetching || containersQuery.isFetching}
        >
          <RefreshCw
            className={
              hostsQuery.isFetching || containersQuery.isFetching
                ? 'h-4 w-4 animate-spin'
                : 'h-4 w-4'
            }
            aria-hidden
          />
          Refresh
        </Button>
      </section>

      <section className="space-y-4" aria-labelledby="hosts-heading">
        <div className="flex items-center justify-between gap-3">
          <h2 id="hosts-heading" className="text-lg font-medium tracking-tight">
            Hosts
          </h2>
          <p className="text-xs text-muted-foreground">
            {hosts.length} registered · {onlineCount} online
          </p>
        </div>

        {hostsQuery.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading hosts…</p>
        ) : hostsQuery.isError ? (
          <ErrorNotice error={hostsQuery.error} label="hosts" />
        ) : hosts.length === 0 ? (
          <EmptyHosts />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {hosts.map((host) => (
              <HostCard
                key={host.id}
                host={host}
                selected={host.id === selectedHostId}
                onSelect={setSelectedHostId}
              />
            ))}
          </div>
        )}
      </section>

      <section className="space-y-4" aria-labelledby="containers-heading">
        <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
          <h2
            id="containers-heading"
            className="text-lg font-medium tracking-tight"
          >
            Containers
          </h2>
          {selectedHost ? (
            <p className="text-xs text-muted-foreground">
              Showing discovery for{' '}
              <span className="text-foreground">{selectedHost.hostname}</span>
            </p>
          ) : null}
        </div>

        {!selectedHostId ? (
          <div className="rounded-lg border border-dashed border-border px-4 py-10 text-center text-sm text-muted-foreground">
            Select a host to view containers.
          </div>
        ) : containersQuery.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading containers…</p>
        ) : containersQuery.isError ? (
          <ErrorNotice error={containersQuery.error} label="containers" />
        ) : (
          <ContainerTable
            containers={containersQuery.data?.containers ?? []}
            busyKey={busyKey}
            onAction={handleContainerAction}
            onViewLogs={setLogsContainer}
          />
        )}

        {selectedHostId && logsContainer ? (
          <ContainerLogsPanel
            hostId={selectedHostId}
            container={logsContainer}
            onClose={() => setLogsContainer(null)}
          />
        ) : null}
      </section>
    </div>
  )
}

function EmptyHosts() {
  return (
    <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border px-4 py-12 text-center">
      <Container className="h-8 w-8 text-muted-foreground" aria-hidden />
      <div className="space-y-1">
        <p className="font-medium">No hosts connected</p>
        <p className="max-w-md text-sm text-muted-foreground">
          Start the DockSight agent against this server. Registered agents
          appear here as Docker hosts.
        </p>
      </div>
    </div>
  )
}

function ErrorNotice({ error, label }: { error: unknown; label: string }) {
  const message =
    error instanceof ApiError
      ? error.message
      : error instanceof Error
        ? error.message
        : `Failed to load ${label}`

  return (
    <div className="rounded-lg border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
      {message}
    </div>
  )
}
