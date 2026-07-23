import { apiClient } from '@/services/api'
import type { Host, HostContainersResponse } from '@/types/api'

export function fetchHosts(): Promise<Host[]> {
  return apiClient.get<Host[]>('/hosts')
}

export function fetchHostContainers(
  hostId: string,
): Promise<HostContainersResponse> {
  return apiClient.get<HostContainersResponse>(`/hosts/${hostId}/containers`)
}
