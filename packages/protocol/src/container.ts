import type { MessageEnvelope } from './envelope'

/**
 * Container message type constants (`domain.action`).
 */
export const CONTAINER_MESSAGE_TYPE = {
  CONTAINER_LIST: 'container.list',
  CONTAINER_LISTED: 'container.listed',
} as const

export type ContainerMessageType =
  (typeof CONTAINER_MESSAGE_TYPE)[keyof typeof CONTAINER_MESSAGE_TYPE]

export const CONTAINER_LIST = CONTAINER_MESSAGE_TYPE.CONTAINER_LIST
export const CONTAINER_LISTED = CONTAINER_MESSAGE_TYPE.CONTAINER_LISTED

/**
 * Payload for `container.list` (Server → Agent).
 * Empty object for now; filters may be added later.
 */
export type ContainerListPayload = Record<string, never> | Record<string, unknown>

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

export type ContainerListMessage = MessageEnvelope<
  typeof CONTAINER_LIST,
  ContainerListPayload
>

export type ContainerListedMessage = MessageEnvelope<
  typeof CONTAINER_LISTED,
  ContainerListedPayload
>

export type ContainerMessage = ContainerListMessage | ContainerListedMessage
