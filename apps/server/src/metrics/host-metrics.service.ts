import { Injectable } from '@nestjs/common';
import type {
  HostCpuMetrics,
  HostMemoryMetrics,
  HostMetricsPayload,
} from '@docksight/protocol';

export type HostMetricsSnapshot = {
  hostId: string;
  agentUuid: string;
  cpu: HostCpuMetrics;
  memory: HostMemoryMetrics;
  /** When the agent sampled the host. */
  collectedAt: Date;
  /** When the server received the sample. */
  receivedAt: Date;
};

/**
 * Latest host CPU/memory reading pushed by each agent over `metrics.host`.
 * In-memory only — nothing is persisted in this increment, so history and
 * cross-restart continuity are not available.
 */
@Injectable()
export class HostMetricsService {
  private readonly byHostId = new Map<string, HostMetricsSnapshot>();
  private readonly hostIdByUuid = new Map<string, string>();

  rememberHost(hostId: string, agentUuid: string) {
    this.hostIdByUuid.set(agentUuid, hostId);
  }

  /**
   * Store an agent sample. Returns null when the agent uuid cannot be mapped to
   * a host id yet (metrics arrived before registration was recorded).
   */
  set(
    agentUuid: string,
    payload: HostMetricsPayload,
    hostId?: string,
  ): HostMetricsSnapshot | null {
    const resolvedHostId = hostId ?? this.hostIdByUuid.get(agentUuid);
    if (!resolvedHostId) {
      return null;
    }

    const snapshot: HostMetricsSnapshot = {
      hostId: resolvedHostId,
      agentUuid,
      cpu: payload.cpu,
      memory: payload.memory,
      collectedAt: parseTimestamp(payload.collectedAt),
      receivedAt: new Date(),
    };

    this.byHostId.set(resolvedHostId, snapshot);
    this.hostIdByUuid.set(agentUuid, resolvedHostId);
    return snapshot;
  }

  getByHostId(hostId: string): HostMetricsSnapshot | null {
    return this.byHostId.get(hostId) ?? null;
  }

  getByUuid(agentUuid: string): HostMetricsSnapshot | null {
    const hostId = this.hostIdByUuid.get(agentUuid);
    if (!hostId) {
      return null;
    }
    return this.byHostId.get(hostId) ?? null;
  }
}

/**
 * `collectedAt` arrives as an untrusted string; fall back to receive time so a
 * malformed value never surfaces as `Invalid Date` in an API response.
 */
function parseTimestamp(value: string | undefined): Date {
  if (!value) {
    return new Date();
  }
  const parsed = new Date(value);
  return Number.isNaN(parsed.valueOf()) ? new Date() : parsed;
}
