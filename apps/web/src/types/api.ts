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

export type ContainerAction = 'start' | 'stop' | 'restart'

export type ContainerActionResult = {
  requestId: string
  action: ContainerAction
  containerId: string
  ok: boolean
  message: string
  error: string | null
}

export type ContainerPort = {
  private: number
  public: string
  protocol: string
}

export type ContainerMount = {
  source: string
  target: string
  mode: string
}

export type ContainerStateDetails = {
  status: string
  running: boolean
  paused: boolean
  restarting: boolean
}

export type ContainerNetwork = {
  name: string
  ip: string
}

export type ContainerInspect = {
  id: string
  shortId: string
  name: string
  image: string
  state: ContainerStateDetails
  created: string
  startedAt: string
  ports: ContainerPort[]
  mounts: ContainerMount[]
  networks: ContainerNetwork[]
  workingDir: string
  cmd: string[]
  restartPolicy: string
}

export type ContainerInspectResult = {
  requestId: string
  container: ContainerInspect | null
  ok: boolean
  error: string | null
}

export type LogEntry = {
  timestamp: string
  stream: string
  message: string
}
