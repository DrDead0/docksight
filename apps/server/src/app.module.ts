import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { databaseConfig } from './common/database/database.config';
import { PrismaModule } from './common/database/prisma.module';
import { RedisModule } from './common/redis/redis.module';
import { WebsocketModule } from './common/websocket/websocket.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: ['.env', '../../.env'],
      load: [databaseConfig],
    }),
    PrismaModule,
    RedisModule,
    WebsocketModule,
  ],
})
export class AppModule {}
