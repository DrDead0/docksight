import { useState } from 'react'
import { useToast } from '@/components/ToastProvider'
import { useContainerAction } from '@/hooks/useContainerAction'
import { ApiError } from '@/services/api'
import type { ContainerAction } from '@/types/api'
import type { ContainerRow } from '@/components/ContainerTable'

const PAST_TENSE: Record<ContainerAction, string> = {
  start: 'started',
  stop: 'stopped',
  restart: 'restarted',
}

/**
 * Runs a lifecycle command against `POST /containers/:id/{action}` and reports
 * the agent's reply as a toast. `busyKey` is `"<containerId>:<action>"`.
 */
export function useContainerCommands(
  fallbackHostId?: string,
  onSettled?: () => void,
) {
  const toast = useToast()
  const mutation = useContainerAction()
  const [busyKey, setBusyKey] = useState<string | null>(null)

  async function run(container: ContainerRow, action: ContainerAction) {
    const hostId = container.hostId ?? fallbackHostId
    if (!hostId) {
      return
    }

    setBusyKey(`${container.id}:${action}`)
    try {
      const result = await mutation.mutateAsync({
        containerId: container.id,
        hostId,
        action,
        containerName: container.name,
      })

      if (result.ok) {
        toast.push({
          tone: 'success',
          title: `Container ${PAST_TENSE[action]}`,
          description: `${container.name.replace(/^\//, '')} · ${result.message}`,
        })
        onSettled?.()
      } else {
        toast.push({
          tone: 'error',
          title: `Could not ${action} container`,
          description: result.error ?? result.message,
        })
      }
    } catch (error) {
      toast.push({
        tone: 'error',
        title: `Could not ${action} container`,
        description:
          error instanceof ApiError || error instanceof Error
            ? error.message
            : `Failed to ${action} container`,
      })
    } finally {
      setBusyKey(null)
    }
  }

  return { busyKey, run }
}
