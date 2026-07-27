import { Logger } from '@nestjs/common';
import {
  OnGatewayConnection,
  OnGatewayDisconnect,
  WebSocketGateway,
} from '@nestjs/websockets';
import { randomUUID } from 'crypto';
import type { IncomingMessage } from 'http';
import type { RawData } from 'ws';
import WebSocket from 'ws';
import {
  AGENT_HEARTBEAT,
  AGENT_REGISTER,
  AGENT_REGISTERED,
  CONTAINER_LIST,
  CONTAINER_LISTED,
  CONTA
  CONTAINER_RESTART,
  CONTAINER_RESULT,
  CONTAINER_START,
  CONTAINER_STOP,
  LOGS_CHUNK,
  LOGS_SUBSCRIBE,
  LOGS_UNSUBSCRIBE,
  createEnvelope,
  isMessageEnvelope,
  type AgentHeartbeatPayload,
  type AgentRegisterPayload,
  type ContainerAction,
  type ContainerListedPayload,
  type ContainerResultPayload,
  type ContainerSummary,
  type LogsChunkPayload,
  type MessageEnvelope,
} from '@docksight/protocol';
import { AgentsService } from './agents.service';
import { ContainerInventoryService } from './container-inventory.service';

type AgentSocket = WebSocket & {
  agentUuid?: string;
  agentId?: string;
};

type PendingList = {
  resolve: (containers: ContainerSummary[]) => void;
  timer: ReturnType<typeof setTimeout>;
};

type PendingCommand = {
  agentUuid: string;
  resolve: (result: ContainerResultPayload) => void;
  reject: (error: Error) => void;
  timer: ReturnType<typeof setTimeout>;
};

type LogSubscriber = {
  agentUuid: string;
  containerId: string;
  onChunk: (payload: LogsChunkPayload) => void;
};

const COMMAND_TYPE: Record<ContainerAction, string> = {
  start: CONTAINER_START,
  stop: CONTAINER_STOP,
  restart: CONTAINER_RESTART,
};

@WebSocketGateway({ path: '/agents' })
export class AgentsGateway
  implements OnGatewayConnection, OnGatewayDisconnect
{
  private readonly logger = new Logger(AgentsGateway.name);
  private readonly socketsByUuid = new Map<string, AgentSocket>();
  private readonly pendingLists = new Map<string, PendingList>();
  private readonly pendingCommands = new Map<string, PendingCommand>();
  private readonly logSubscribers = new Map<string, LogSubscriber>();

  constructor(
    private readonly agentsService: AgentsService,
    private readonly inventory: ContainerInventoryService,
  ) {}

  handleConnection(client: AgentSocket, ...args: unknown[]) {
    const request = args[0] as IncomingMessage | undefined;
    this.logger.log(
      `Agent connected from ${request?.socket.remoteAddress ?? 'unknown'}`,
    );

    client.on('message', (data: RawData) => {
      void this.handleMessage(client, data);
    });

    client.on('error', (error) => {
      this.logger.warn(`Agent socket error: ${error.message}`);
    });
  }

  async handleDisconnect(client: AgentSocket) {
    const uuid = client.agentUuid;
    this.logger.log(`Agent disconnected uuid=${uuid ?? 'unknown'}`);
    if (uuid) {
      this.socketsByUuid.delete(uuid);
      const pending = this.pendingLists.get(uuid);
      if (pending) {
        clearTimeout(pending.timer);
        this.pendingLists.delete(uuid);
        pending.resolve(this.inventory.getByUuid(uuid)?.containers ?? []);
      }
      this.rejectPendingCommandsForAgent(
        uuid,
        new Error('Agent disconnected before command completed'),
      );
      this.unsubscribeLogsForAgent(uuid);
      await this.agentsService.markOffline(uuid);
    }
  }

  /**
   * Ask a connected agent for a fresh container list, or return the last cache.
   */
  async requestContainerList(
    agentUuid: string,
    hostId: string,
    timeoutMs = 5_000,
  ): Promise<ContainerSummary[]> {
    this.inventory.rememberHost(hostId, agentUuid);

    const client = this.socketsByUuid.get(agentUuid);
    if (!client || client.readyState !== WebSocket.OPEN) {
      return this.inventory.getByHostId(hostId)?.containers ?? [];
    }

    return new Promise<ContainerSummary[]>((resolve) => {
      const previous = this.pendingLists.get(agentUuid);
      if (previous) {
        clearTimeout(previous.timer);
        this.pendingLists.delete(agentUuid);
        previous.resolve(
          this.inventory.getByUuid(agentUuid)?.containers ?? [],
        );
      }

      const timer = setTimeout(() => {
        this.pendingLists.delete(agentUuid);
        resolve(this.inventory.getByHostId(hostId)?.containers ?? []);
      }, timeoutMs);

      this.pendingLists.set(agentUuid, { resolve, timer });
      this.send(client, createEnvelope(CONTAINER_LIST, {}));
      this.logger.log(`Requested container.list from uuid=${agentUuid}`);
    });
  }

  /**
   * Forward a lifecycle command to a connected agent and wait for container.result.
   */
  async sendContainerCommand(
    agentUuid: string,
    hostId: string,
    action: ContainerAction,
    containerId: string,
    timeoutMs = 45_000,
  ): Promise<ContainerResultPayload> {
    this.inventory.rememberHost(hostId, agentUuid);

    const client = this.socketsByUuid.get(agentUuid);
    if (!client || client.readyState !== WebSocket.OPEN) {
      throw new Error('Agent is not connected');
    }

    const requestId = randomUUID();
    const type = COMMAND_TYPE[action];

    return new Promise<ContainerResultPayload>((resolve, reject) => {
      const timer = setTimeout(() => {
        this.pendingCommands.delete(requestId);
        reject(
          new Error(
            `Agent timed out waiting for container.result (${action})`,
          ),
        );
      }, timeoutMs);

      this.pendingCommands.set(requestId, {
        agentUuid,
        resolve,
        reject,
        timer,
      });

      this.send(
        client,
        createEnvelope(type, {
          requestId,
          containerId,
        }),
      );

      this.logger.log(
        `Sent ${type} requestId=${requestId} containerId=${containerId.slice(0, 12)} uuid=${agentUuid}`,
      );
    });
  }

  /**
   * Ask an agent to stream container logs; chunks are delivered via onChunk.
   * Returns the correlation requestId used for unsubscribe.
   */
  startLogStream(options: {
    agentUuid: string;
    hostId: string;
    containerId: string;
    tail?: number;
    follow?: boolean;
    onChunk: (payload: LogsChunkPayload) => void;
  }): string {
    const {
      agentUuid,
      hostId,
      containerId,
      tail = 100,
      follow = true,
      onChunk,
    } = options;

    this.inventory.rememberHost(hostId, agentUuid);

    const client = this.socketsByUuid.get(agentUuid);
    if (!client || client.readyState !== WebSocket.OPEN) {
      throw new Error('Agent is not connected');
    }

    const requestId = randomUUID();
    this.logSubscribers.set(requestId, {
      agentUuid,
      containerId,
      onChunk,
    });

    this.send(
      client,
      createEnvelope(LOGS_SUBSCRIBE, {
        requestId,
        containerId,
        tail,
        follow,
      }),
    );

    this.logger.log(
      `Sent logs.subscribe requestId=${requestId} containerId=${containerId.slice(0, 12)} uuid=${agentUuid}`,
    );

    return requestId;
  }

  stopLogStream(requestId: string) {
    const subscriber = this.logSubscribers.get(requestId);
    if (!subscriber) {
      return;
    }
    this.logSubscribers.delete(requestId);

    const client = this.socketsByUuid.get(subscriber.agentUuid);
    if (client && client.readyState === WebSocket.OPEN) {
      this.send(
        client,
        createEnvelope(LOGS_UNSUBSCRIBE, { requestId }),
      );
    }

    this.logger.log(`Sent logs.unsubscribe requestId=${requestId}`);
  }

  private unsubscribeLogsForAgent(agentUuid: string) {
    for (const [requestId, subscriber] of this.logSubscribers.entries()) {
      if (subscriber.agentUuid !== agentUuid) {
        continue;
      }
      this.logSubscribers.delete(requestId);
    }
  }

  private rejectPendingCommandsForAgent(agentUuid: string, error: Error) {
    for (const [requestId, pending] of this.pendingCommands.entries()) {
      if (pending.agentUuid !== agentUuid) {
        continue;
      }
      clearTimeout(pending.timer);
      this.pendingCommands.delete(requestId);
      pending.reject(error);
    }
  }

  private async handleMessage(client: AgentSocket, data: RawData) {
    let parsed: unknown;
    try {
      parsed = JSON.parse(data.toString()) as unknown;
    } catch {
      this.send(
        client,
        createEnvelope(AGENT_REGISTERED, {
          id: '',
          uuid: '',
          status: 'UNKNOWN',
          message: 'Invalid JSON message',
        }),
      );
      return;
    }

    if (!isMessageEnvelope(parsed)) {
      this.logger.warn('Ignored non-envelope agent message');
      return;
    }

    const envelope = parsed;

    switch (envelope.type) {
      case AGENT_REGISTER:
        await this.handleRegister(
          client,
          envelope.payload as AgentRegisterPayload,
        );
        break;
      case AGENT_HEARTBEAT:
        await this.handleHeartbeat(
          client,
          envelope.payload as AgentHeartbeatPayload,
        );
        break;
      case CONTAINER_LISTED:
        this.handleContainerListed(
          client,
          envelope.payload as ContainerListedPayload,
        );
        break;
      case CONTAINER_RESULT:
        this.handleContainerResult(
          envelope.payload as ContainerResultPayload,
        );
        break;
      case LOGS_CHUNK:
        this.handleLogsChunk(envelope.payload as LogsChunkPayload);
        break;
      default:
        this.logger.warn(`Unknown agent message type: ${envelope.type}`);
    }
  }

  private async handleRegister(
    client: AgentSocket,
    payload: AgentRegisterPayload,
  ) {
    if (!payload?.uuid || !payload.hostname || !payload.os) {
      this.send(
        client,
        createEnvelope(AGENT_REGISTERED, {
          id: '',
          uuid: payload?.uuid ?? '',
          status: 'UNKNOWN',
          message: 'Invalid registration payload',
        }),
      );
      return;
    }

    const result = await this.agentsService.register({
      uuid: payload.uuid,
      hostname: payload.hostname,
      os: payload.os,
      architecture: payload.architecture ?? 'unknown',
      version: payload.version ?? 'unknown',
    });

    client.agentUuid = result.uuid;
    client.agentId = result.id;
    this.socketsByUuid.set(result.uuid, client);
    this.inventory.rememberHost(result.id, result.uuid);

    this.logger.log(
      `Agent registered: hostname=${payload.hostname} uuid=${result.uuid}`,
    );

    this.send(client, createEnvelope(AGENT_REGISTERED, result));

    this.send(client, createEnvelope(CONTAINER_LIST, {}));
    this.logger.log(`Sent container.list to agent uuid=${result.uuid}`);
  }

  private async handleHeartbeat(
    client: AgentSocket,
    payload: AgentHeartbeatPayload,
  ) {
    if (!payload?.uuid) {
      return;
    }

    client.agentUuid = payload.uuid;
    this.socketsByUuid.set(payload.uuid, client);
    await this.agentsService.heartbeat(payload.uuid);
  }

  private handleContainerListed(
    client: AgentSocket,
    payload: ContainerListedPayload,
  ) {
    const containers = Array.isArray(payload?.containers)
      ? payload.containers
      : [];
    const uuid = client.agentUuid;

    this.logger.log(
      `Received container.listed from uuid=${uuid ?? 'unknown'} count=${containers.length}`,
    );

    if (uuid) {
      this.inventory.setContainers(uuid, containers, client.agentId);
      const pending = this.pendingLists.get(uuid);
      if (pending) {
        clearTimeout(pending.timer);
        this.pendingLists.delete(uuid);
        pending.resolve(containers);
      }
    }
  }

  private handleContainerResult(payload: ContainerResultPayload) {
    if (!payload?.requestId) {
      this.logger.warn('Ignored container.result without requestId');
      return;
    }

    const pending = this.pendingCommands.get(payload.requestId);
    if (!pending) {
      this.logger.warn(
        `No pending command for container.result requestId=${payload.requestId}`,
      );
      return;
    }

    clearTimeout(pending.timer);
    this.pendingCommands.delete(payload.requestId);

    this.logger.log(
      `Received container.result requestId=${payload.requestId} action=${payload.action} ok=${payload.ok}`,
    );

    pending.resolve({
      requestId: payload.requestId,
      action: payload.action,
      containerId: payload.containerId,
      ok: Boolean(payload.ok),
      message: payload.message ?? '',
      error: payload.error ?? null,
    });
  }

  private handleLogsChunk(payload: LogsChunkPayload) {
    if (!payload?.requestId) {
      this.logger.warn('Ignored logs.chunk without requestId');
      return;
    }

    const subscriber = this.logSubscribers.get(payload.requestId);
    if (!subscriber) {
      this.logger.warn(
        `No subscriber for logs.chunk requestId=${payload.requestId}`,
      );
      return;
    }

    subscriber.onChunk({
      requestId: payload.requestId,
      containerId: payload.containerId,
      entries: Array.isArray(payload.entries) ? payload.entries : [],
    });
  }

  private send(client: WebSocket, envelope: MessageEnvelope) {
    if (client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify(envelope));
    }
  }
}
