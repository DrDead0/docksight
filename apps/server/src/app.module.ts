import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { APP_GUARD } from '@nestjs/core';
import { AgentsModule } from './agents/agents.module';
import { JwtAuthGuard } from './auth/jwt-auth.guard';
import { RolesGuard } from './auth/roles.guard';
import { authConfig } from './auth/auth.config';
import { AuthModule } from './auth/auth.module';
import { databaseConfig } from './common/database/database.config';
import { PrismaModule } from './common/database/prisma.module';
import { RedisModule } from './common/redis/redis.module';
import { ContainersModule } from './containers/containers.module';
import { HostsModule } from './hosts/hosts.module';
import { SetupModule } from './setup/setup.module';
import { UsersModule } from './users/users.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: ['.env', '../../.env'],
      // authConfig is loaded here as well as in AuthModule so a missing or
      // too-short JWT_SECRET aborts startup instead of failing at first login.
      load: [databaseConfig, authConfig],
    }),
    PrismaModule,
    RedisModule,
    AgentsModule,
    HostsModule,
    ContainersModule,
    UsersModule,
    AuthModule,
    SetupModule,
  ],
  providers: [
    // Authentication is global and fail-closed: every HTTP route requires a
    // valid token unless it is explicitly marked @Public(). A new controller
    // is protected the day it is written, with no decorator to forget.
    //
    // Order matters — Nest runs global guards in registration order, so
    // JwtAuthGuard populates request.user before RolesGuard reads it.
    { provide: APP_GUARD, useClass: JwtAuthGuard },
    { provide: APP_GUARD, useClass: RolesGuard },
  ],
})
export class AppModule {}
