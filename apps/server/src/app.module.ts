import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { AuthModule } from './auth/auth.module';
import { UsersModule } from './users/users.module';
import { HostsModule } from './hosts/hosts.module';
import { AgentsModule } from './agents/agents.module';
import { ContainersModule } from './containers/containers.module';
import { LogsModule } from './logs/logs.module';
import { MetricsModule } from './metrics/metrics.module';
import { EnvironmentsModule } from './environments/environments.module';
import { NotificationsModule } from './notifications/notifications.module';
import { AuditModule } from './audit/audit.module';
import { PrismaModule } from './common/prisma/prisma.module';
import { RedisModule } from './common/redis/redis.module';
import { HealthModule } from './common/health/health.module';
import { WebsocketModule } from './common/websocket/websocket.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: ['.env', '../../.env'],
    }),
    PrismaModule,
    RedisModule,
    WebsocketModule,
    HealthModule,
    AuthModule,
    UsersModule,
    HostsModule,
    AgentsModule,
    ContainersModule,
    LogsModule,
    MetricsModule,
    EnvironmentsModule,
    NotificationsModule,
    AuditModule,
  ],
})
export class AppModule {}
