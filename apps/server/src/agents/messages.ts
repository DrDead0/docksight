export type AgentMessageType =
  | 'agent.register'
  | 'agent.registered'
  | 'agent.heartbeat';

export type AgentMessageEnvelope<T = unknown> = {
  type: AgentMessageType;
  payload: T;
};

export type AgentRegisterPayload = {
  uuid: string;
  hostname: string;
  os: string;
  architecture: string;
  version: string;
};

export type AgentRegisteredPayload = {
  id: string;
  uuid: string;
  status: string;
  message: string;
};

export type AgentHeartbeatPayload = {
  uuid: string;
};
