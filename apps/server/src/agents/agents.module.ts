import { Module } from '@nestjs/common';
import { AgentsGateway } from './agents.gateway';
import { AgentsService } from './agents.service';

@Module({
  providers: [AgentsService, AgentsGateway],
  exports: [AgentsService],
})
export class AgentsModule {}
