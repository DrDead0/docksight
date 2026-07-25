/**
 * @docksight/protocol
 *
 * Shared DockSight agent WebSocket protocol contracts.
 * See package README and docs/protocol.md for the full specification.
 */

export type { JsonObject, MessageType, ProtocolDomain } from './types'

export type { MessageEnvelope } from './envelope'
export { createEnvelope, isJsonObject, isMessageEnvelope } from './envelope'

export type {
  AgentHeartbeatMessage,
  AgentHeartbeatPayload,
  AgentMessage,
  AgentMessageType,
  AgentRegisterMessage,
  AgentRegisterPayload,
  AgentRegisteredMessage,
  AgentRegisteredPayload,
  AgentStatus,
} from './agent'

export {
  AGENT_HEARTBEAT,
  AGENT_MESSAGE_TYPE,
  AGENT_REGISTER,
  AGENT_REGISTERED,
  AGENT_STATUS,
} from './agent'

export type {
  ContainerAction,
  ContainerCommandPayload,
  ContainerListMessage,
  ContainerListPayload,
  ContainerListedMessage,
  ContainerListedPayload,
  ContainerMessage,
  ContainerMessageType,
  ContainerRestartMessage,
  ContainerResultMessage,
  ContainerResultPayload,
  ContainerStartMessage,
  ContainerStopMessage,
  ContainerSummary,
} from './container'

export {
  CONTAINER_LIST,
  CONTAINER_LISTED,
  CONTAINER_MESSAGE_TYPE,
  CONTAINER_RESTART,
  CONTAINER_RESULT,
  CONTAINER_START,
  CONTAINER_STOP,
} from './container'

export type { ProtocolErrorCode } from './errors'
export { ERROR_DOMAIN, PROTOCOL_ERROR_CODE } from './errors'
