import { useMutation, useQueryClient } from '@tanstack/react-query'
import { containersQueryKey } from '@/hooks/useContainers'
import { runContainerAction } from '@/services/hosts'
import type { ContainerAction, ContainerActionResult } from '@/types/api'

type ContainerActionVariables = {
  containerId: string
  hostId: string
  action: ContainerAction
  containerName?: string
}

export function useContainerAction() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      containerId,
      hostId,
      action,
    }: ContainerActionVariables): Promise<ContainerActionResult> =>
      runContainerAction(containerId, hostId, action),
    onSuccess: async (_result, variables) => {
      await queryClient.invalidateQueries({
        queryKey: containersQueryKey(variables.hostId),
      })
    },
  })
}
