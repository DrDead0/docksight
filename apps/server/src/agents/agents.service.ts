import { Injectable, Logger } from '@nestjs/common';
import type {
  AgentRegisterPayload,
  AgentRegisteredPayload,
} from '@docksight/protocol';
import { Agent, AgentStatus } from '../../generated/prisma/client';
import { PrismaService } from '../common/database/prisma.service';

@Injectable()
export class AgentsService {
  private readonly logger = new Logger(AgentsService.name);

  constructor(private readonly prisma: PrismaService) {}

  async register(payload: AgentRegisterPayload): Promise<AgentRegisteredPayload> {
    const now = new Date();

    const agent = await this.prisma.agent.upsert({
      where: { uuid: payload.uuid },
      create: {
        uuid: payload.uuid,
        hostname: payload.hostname,
        os: payload.os,
        architecture: payload.architecture,
        version: payload.version,
        status: AgentStatus.ONLINE,
        lastSeen: now,
      },
      update: {
        hostname: payload.hostname,
        os: payload.os,
        architecture: payload.architecture,
        version: payload.version,
        status: AgentStatus.ONLINE,
        lastSeen: now,
      },
    });

    this.logger.log(
      `Agent registered uuid=${agent.uuid} hostname=${agent.hostname} id=${agent.id}`,
    );

    return {
      id: agent.id,
      uuid: agent.uuid,
      status: agent.status,
      message: 'Registration successful',
    };
  }

  async heartbeat(uuid: string): Promise<Agent | null> {
    const existing = await this.prisma.agent.findUnique({ where: { uuid } });
    if (!existing) {
      this.logger.warn(`Heartbeat from unknown agent uuid=${uuid}`);
      return null;
    }

    return this.prisma.agent.update({
      where: { uuid },
      data: {
        status: AgentStatus.ONLINE,
        lastSeen: new Date(),
      },
    });
  }

  async markOffline(uuid: string): Promise<void> {
    try {
      await this.prisma.agent.update({
        where: { uuid },
        data: { status: AgentStatus.OFFLINE },
      });
    } catch {
      // Agent may not exist yet if disconnect happens before registration.
    }
  }

  findByUuid(uuid: string): Promise<Agent | null> {
    return this.prisma.agent.findUnique({ where: { uuid } });
  }

  findById(id: string): Promise<Agent | null> {
    return this.prisma.agent.findUnique({ where: { id } });
  }

  findAll(): Promise<Agent[]> {
    return this.prisma.agent.findMany({ orderBy: { createdAt: 'desc' } });
  }
}
