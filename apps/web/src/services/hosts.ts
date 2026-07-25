import { apiClient } from '@/services/api'
import type {
  ContainerAction,
  ContainerActionResult,
  Host,
  HostContainersResponse,
} from '@/types/api'

export function fetchHosts(): Promise<Host[]> {
  return apiClient.get<Host[]>('/hosts')
}

export function fetchHostContainers(
  hostId: string,
): Promise<HostContainersResponse> {
  return apiClient.get<HostContainersResponse>(`/hosts/${hostId}/containers`)
}

export function runContainerAction(
  containerId: string,
  hostId: string,
  action: ContainerAction,
): Promise<ContainerActionResult> {
  return apiClient.post<ContainerActionResult>(
    `/containers/${encodeURIComponent(containerId)}/${action}`,
    { hostId },
  )
}
