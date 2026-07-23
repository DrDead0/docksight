export type HostStatus = 'ONLINE' | 'OFFLINE' | 'UNKNOWN' | string

export type Host = {
  id: string
  uuid: string
  hostname: string
  os: string
  architecture: string
  version: string
  status: HostStatus
  lastSeen: string | null
}

export type Container = {
  id: string
  name: string
  image: string
  status: string
  state: string
}

export type HostContainersResponse = {
  hostId: string
  containers: Container[]
  updatedAt: string | null
}
