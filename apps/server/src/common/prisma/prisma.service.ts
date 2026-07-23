import { Injectable, Logger, OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import { PrismaClient } from '@prisma/client';

@Injectable()
export class PrismaService
  extends PrismaClient
  implements OnModuleInit, OnModuleDestroy
{
  private readonly logger = new Logger(PrismaService.name);

  async onModuleInit() {
    try {
      await this.$connect();
      this.logger.log('PostgreSQL connected');
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      this.logger.warn(
        `PostgreSQL unavailable (${message}). Start infra with: npm run docker:infra`,
      );
    }
  }

  async onModuleDestroy() {
    await this.$disconnect();
  }
}
