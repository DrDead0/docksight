import type { MessageEnvelope } from './envelope'

/**
 * Container message type constants (`domain.action`).
 */
export const CONTAINER_MESSAGE_TYPE = {
  CONTAINER_LIST: 'container.list',
  CONTAINER_LISTED: 'container.listed',
  CONTAINER_START: 'container.start',
  CONTAINER_STOP: 'container.stop',
  CONTAINER_RESTART: 'container.restart',
  CONTAINER_RESULT: 'container.result',
} as const

export type ContainerMessageType =
  (typeof CONTAINER_MESSAGE_TYPE)[keyof typeof CONTAINER_MESSAGE_TYPE]

export const CONTAINER_LIST = CONTAINER_MESSAGE_TYPE.CONTAINER_LIST
export const CONTAINER_LISTED = CONTAINER_MESSAGE_TYPE.CONTAINER_LISTED
export const CONTAINER_START = CONTAINER_MESSAGE_TYPE.CONTAINER_START
export const CONTAINER_STOP = CONTAINER_MESSAGE_TYPE.CONTAINER_STOP
export const CONTAINER_RESTART = CONTAINER_MESSAGE_TYPE.CONTAINER_RESTART
export const CONTAINER_RESULT = CONTAINER_MESSAGE_TYPE.CONTAINER_RESULT

/**
 * Payload for `container.list` (Server → Agent).
 * Empty object for now; filters may be added later.
 */
export type ContainerListPayload =
  | Record<string, never>
  | Record<string, unknown>

/**
 * One container summary returned by discovery.
 */
export type ContainerSummary = {
  id: string
  name: string
  image: string
  status: string
  state: string
}

/**
 * Payload for `container.listed` (Agent → Server).
 */
export type ContainerListedPayload = {
  containers: ContainerSummary[]
}

/**
 * Lifecycle actions that produce `container.result`.
 */
export type ContainerAction = 'start' | 'stop' | 'restart'

/**
 * Shared command payload for start / stop / restart (Server → Agent).
 */
export type ContainerCommandPayload = {
  requestId: string
  containerId: string
}

/**
 * Payload for `container.result` (Agent → Server).
 */
export type ContainerResultPayload = {
  requestId: string
  action: ContainerAction
  containerId: string
  ok: boolean
  message: string
  error: string | null
}

export type ContainerListMessage = MessageEnvelope<
  typeof CONTAINER_LIST,
  ContainerListPayload
>

export type ContainerListedMessage = MessageEnvelope<
  typeof CONTAINER_LISTED,
  ContainerListedPayload
>

export type ContainerStartMessage = MessageEnvelope<
  typeof CONTAINER_START,
  ContainerCommandPayload
>

export type ContainerStopMessage = MessageEnvelope<
  typeof CONTAINER_STOP,
  ContainerCommandPayload
>

export type ContainerRestartMessage = MessageEnvelope<
  typeof CONTAINER_RESTART,
  ContainerCommandPayload
>

export type ContainerResultMessage = MessageEnvelope<
  typeof CONTAINER_RESULT,
  ContainerResultPayload
>

export type ContainerMessage =
  | ContainerListMessage
  | ContainerListedMessage
  | ContainerStartMessage
  | ContainerStopMessage
  | ContainerRestartMessage
  | ContainerResultMessage
