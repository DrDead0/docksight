import { Injectable } from '@nestjs/common';
import type { ContainerSummary } from '@docksight/protocol';
import { AgentsGateway } from '../agents/agents.gateway';
import { AgentsService } from '../agents/agents.service';
import { ContainerInventoryService } from '../agents/container-inventory.service';

export type HostDto = {
  id: string;
  uuid: string;
  hostname: string;
  os: string;
  architecture: string;
  version: string;
  status: string;
  lastSeen: string | null;
};

export type HostContainersDto = {
  hostId: string;
  containers: ContainerSummary[];
  updatedAt: string | null;
};

@Injectable()
export class HostsService {
  constructor(
    private readonly agentsService: AgentsService,
    private readonly inventory: ContainerInventoryService,
    private readonly agentsGateway: AgentsGateway,
  ) {}

  async listHosts(): Promise<HostDto[]> {
    const agents = await this.agentsService.findAll();

    for (const agent of agents) {
      this.inventory.rememberHost(agent.id, agent.uuid);
    }

    return agents.map((agent) => ({
      id: agent.id,
      uuid: agent.uuid,
      hostname: agent.hostname,
      os: agent.os,
      architecture: agent.architecture,
      version: agent.version,
      status: agent.status,
      lastSeen: agent.lastSeen ? agent.lastSeen.toISOString() : null,
    }));
  }

  async listContainers(hostId: string): Promise<HostContainersDto | null> {
    const agent = await this.agentsService.findById(hostId);
    if (!agent) {
      return null;
    }

    this.inventory.rememberHost(agent.id, agent.uuid);

    const containers = await this.agentsGateway.requestContainerList(
      agent.uuid,
      agent.id,
    );
    const snapshot = this.inventory.getByHostId(agent.id);

    return {
      hostId: agent.id,
      containers,
      updatedAt: snapshot?.updatedAt
        ? snapshot.updatedAt.toISOString()
        : null,
    };
  }
}
