import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { AgentsModule } from './agents/agents.module';
import { databaseConfig } from './common/database/database.config';
import { PrismaModule } from './common/database/prisma.module';
import { RedisModule } from './common/redis/redis.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: ['.env', '../../.env'],
      load: [databaseConfig],
    }),
    PrismaModule,
    RedisModule,
    AgentsModule,
  ],
})
export class AppModule {}
