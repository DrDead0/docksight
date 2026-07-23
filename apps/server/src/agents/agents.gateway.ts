import { Logger } from '@nestjs/common';
import {
  OnGatewayConnection,
  OnGatewayDisconnect,
  WebSocketGateway,
} from '@nestjs/websockets';
import type { IncomingMessage } from 'http';
import type { RawData } from 'ws';
import WebSocket from 'ws';
import {
  AGENT_HEARTBEAT,
  AGENT_REGISTER,
  AGENT_REGISTERED,
  CONTAINER_LIST,
  CONTAINER_LISTED,
  createEnvelope,
  isMessageEnvelope,
  type AgentHeartbeatPayload,
  type AgentRegisterPayload,
  type ContainerListedPayload,
  type ContainerSummary,
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

@WebSocketGateway({ path: '/agents' })
export class AgentsGateway
  implements OnGatewayConnection, OnGatewayDisconnect
{
  private readonly logger = new Logger(AgentsGateway.name);
  private readonly socketsByUuid = new Map<string, AgentSocket>();
  private readonly pendingLists = new Map<string, PendingList>();

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

    // Kick off read-only container discovery after successful registration.
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

    for (const container of containers) {
      this.logger.log(
        `container id=${container.id?.slice(0, 12) ?? '?'} name=${container.name} image=${container.image} state=${container.state} status=${container.status}`,
      );
    }

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

  private send(client: WebSocket, envelope: MessageEnvelope) {
    if (client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify(envelope));
    }
  }
}
