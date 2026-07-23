/** Shared domain types across DockSight apps. */

export type HealthStatus = {
  status: string
  service: string
  timestamp?: string
}

export type HostId = string
export type ContainerId = string
export type AgentId = string
