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

export type { ProtocolErrorCode } from './errors'
export { ERROR_DOMAIN, PROTOCOL_ERROR_CODE } from './errors'
