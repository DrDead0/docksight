import { Injectable } from '@nestjs/common';
import type {
  ContainerSummary,
  HostCpuMetrics,
  HostMemoryMetrics,
} from '@docksight/protocol';
import { AgentsGateway } from '../agents/agents.gateway';
import { AgentsService } from '../agents/agents.service';
import { ContainerInventoryService } from '../agents/container-inventory.service';
import {
  HostMetricsService,
  type HostMetricsSnapshot,
} from '../metrics/host-metrics.service';

export type HostMetricsDto = {
  hostId: string;
  cpu: HostCpuMetrics | null;
  memory: HostMemoryMetrics | null;
  /** When the agent sampled the host; null until the first sample arrives. */
  collectedAt: string | null;
};

export type HostDto = {
  id: string;
  uuid: string;
  hostname: string;
  os: string;
  architecture: string;
  version: string;
  status: string;
  lastSeen: string | null;
  /**
   * Latest reported usage, so the host list needs no extra round-trips. Always
   * present; its `cpu`/`memory` are null until the agent reports.
   */
  metrics: HostMetricsDto;
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
    private readonly hostMetrics: HostMetricsService,
  ) {}

  async listHosts(): Promise<HostDto[]> {
    const agents = await this.agentsService.findAll();

    for (const agent of agents) {
      this.inventory.rememberHost(agent.id, agent.uuid);
      this.hostMetrics.rememberHost(agent.id, agent.uuid);
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
      metrics: toMetricsDto(agent.id, this.hostMetrics.getByHostId(agent.id)),
    }));
  }

  /**
   * Latest CPU/memory sample for a host. Returns null only when the host is
   * unknown — a known host that has not reported yet yields null fields, so the
   * dashboard can tell "no such host" apart from "no data yet".
   */
  async getMetrics(hostId: string): Promise<HostMetricsDto | null> {
    const agent = await this.agentsService.findById(hostId);
    if (!agent) {
      return null;
    }

    this.hostMetrics.rememberHost(agent.id, agent.uuid);
    return toMetricsDto(agent.id, this.hostMetrics.getByHostId(agent.id));
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
      updatedAt: snapshot?.updatedAt ? snapshot.updatedAt.toISOString() : null,
    };
  }
}

function toMetricsDto(
  hostId: string,
  snapshot: HostMetricsSnapshot | null,
): HostMetricsDto {
  return {
    hostId,
    cpu: snapshot?.cpu ?? null,
    memory: snapshot?.memory ?? null,
    collectedAt: snapshot?.collectedAt.toISOString() ?? null,
  };
}
