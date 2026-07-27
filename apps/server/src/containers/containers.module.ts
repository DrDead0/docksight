import { Module } from '@nestjs/common';
import { AgentsModule } from '../agents/agents.module';
import { ContainersController } from './containers.controller';
import { ContainersService } from './containers.service';

@Module({
  // No AuthModule import needed: the guards are registered globally in
  // AppModule, and @Roles() is inert metadata that requires no injection.
  imports: [AgentsModule],
  controllers: [ContainersController],
  providers: [ContainersService],
})
export class ContainersModule {}
