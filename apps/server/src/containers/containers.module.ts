import { Module } from '@nestjs/common';
import { AgentsModule } from '../agents/agents.module';
import { ContainersController } from './containers.controller';
import { ContainersService } from './containers.service';

@Module({
  imports: [AgentsModule],
  controllers: [ContainersController],
  providers: [ContainersService],
})
export class ContainersModule {}
