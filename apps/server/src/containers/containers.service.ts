import { Injectable } from '@nestjs/common';
import type {
  ContainerAction,
  ContainerResultPayload,
} from '@docksight/protocol';
import { AgentsGateway } from '../agents/agents.gateway';
import { AgentsService } from '../agents/agents.service';

@Injectable()
export class ContainersService {
  constructor(
    private readonly agentsService: AgentsService,
    private readonly agentsGateway: AgentsGateway,
  ) {}

  async runLifecycleAction(
    hostId: string,
    containerId: string,
    action: ContainerAction,
  ): Promise<ContainerResultPayload> {
    if (!hostId?.trim()) {
      throw new Error('hostId is required');
    }
    if (!containerId?.trim()) {
      throw new Error('containerId is required');
    }

    const agent = await this.agentsService.findById(hostId);
    if (!agent) {
      throw new Error(`Host not found: ${hostId}`);
    }

    return this.agentsGateway.sendContainerCommand(
      agent.uuid,
      agent.id,
      action,
      containerId,
    );
  }
}
