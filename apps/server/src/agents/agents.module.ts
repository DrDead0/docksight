import { Module } from '@nestjs/common';
import { AgentsGateway } from './agents.gateway';
import { AgentsService } from './agents.service';
import { ContainerInventoryService } from './container-inventory.service';

@Module({
  providers: [AgentsService, AgentsGateway, ContainerInventoryService],
  exports: [AgentsService, AgentsGateway, ContainerInventoryService],
})
export class AgentsModule {}
