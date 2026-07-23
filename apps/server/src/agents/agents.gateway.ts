import { Logger } from '@nestjs/common';
import {
  OnGatewayConnection,
  OnGatewayDisconnect,
  WebSocketGateway,
} from '@nestjs/websockets';
import type { IncomingMessage } from 'http';
import type { RawData } from 'ws';
import WebSocket from 'ws';
import { AgentsService } from './agents.service';
import type {
  AgentHeartbeatPayload,
  AgentMessageEnvelope,
  AgentRegisterPayload,
} from './messages';

type AgentSocket = WebSocket & {
  agentUuid?: string;
};

@WebSocketGateway({ path: '/agents' })
export class AgentsGateway
  implements OnGatewayConnection, OnGatewayDisconnect
{
  private readonly logger = new Logger(AgentsGateway.name);

  constructor(private readonly agentsService: AgentsService) {}

  handleConnection(client: AgentSocket, ...args: unknown[]) {
    const request = args[0] as IncomingMessage | undefined;
    this.logger.log(
      `Agent socket connected from ${request?.socket.remoteAddress ?? 'unknown'}`,
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
    this.logger.log(`Agent socket disconnected uuid=${uuid ?? 'unknown'}`);
    if (uuid) {
      await this.agentsService.markOffline(uuid);
    }
  }

  private async handleMessage(client: AgentSocket, data: RawData) {
    let envelope: AgentMessageEnvelope;
    try {
      envelope = JSON.parse(data.toString()) as AgentMessageEnvelope;
    } catch {
      this.send(client, {
        type: 'agent.registered',
        payload: {
          id: '',
          uuid: '',
          status: 'UNKNOWN',
          message: 'Invalid JSON message',
        },
      });
      return;
    }

    switch (envelope.type) {
      case 'agent.register':
        await this.handleRegister(
          client,
          envelope.payload as AgentRegisterPayload,
        );
        break;
      case 'agent.heartbeat':
        await this.handleHeartbeat(
          client,
          envelope.payload as AgentHeartbeatPayload,
        );
        break;
      default:
        this.logger.warn(`Unknown agent message type: ${String(envelope.type)}`);
    }
  }

  private async handleRegister(
    client: AgentSocket,
    payload: AgentRegisterPayload,
  ) {
    if (!payload?.uuid || !payload.hostname || !payload.os) {
      this.send(client, {
        type: 'agent.registered',
        payload: {
          id: '',
          uuid: payload?.uuid ?? '',
          status: 'UNKNOWN',
          message: 'Invalid registration payload',
        },
      });
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
    this.send(client, { type: 'agent.registered', payload: result });
  }

  private async handleHeartbeat(
    client: AgentSocket,
    payload: AgentHeartbeatPayload,
  ) {
    if (!payload?.uuid) {
      return;
    }

    client.agentUuid = payload.uuid;
    await this.agentsService.heartbeat(payload.uuid);
  }

  private send(client: WebSocket, envelope: AgentMessageEnvelope) {
    if (client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify(envelope));
    }
  }
}
