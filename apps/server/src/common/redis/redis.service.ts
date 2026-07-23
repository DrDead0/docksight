import { Injectable, Logger, OnModuleDestroy } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import Redis from 'ioredis';

/**
 * Redis connection placeholder.
 * Connection is established lazily; callers should handle unavailable Redis
 * until caching/pub-sub features are implemented.
 */
@Injectable()
export class RedisService implements OnModuleDestroy {
  private readonly logger = new Logger(RedisService.name);
  private readonly client: Redis;

  constructor(private readonly config: ConfigService) {
    const url = this.config.get<string>('REDIS_URL');
    const host = this.config.get<string>('REDIS_HOST', 'localhost');
    const port = this.config.get<number>('REDIS_PORT', 6379);

    this.client = url
      ? new Redis(url, { lazyConnect: true, maxRetriesPerRequest: 1 })
      : new Redis({ host, port, lazyConnect: true, maxRetriesPerRequest: 1 });

    this.client.on('error', (error) => {
      this.logger.warn(`Redis error: ${error.message}`);
    });
  }

  getClient(): Redis {
    return this.client;
  }

  async connect(): Promise<void> {
    if (this.client.status === 'wait') {
      await this.client.connect();
      this.logger.log('Redis connected');
    }
  }

  async onModuleDestroy() {
    if (this.client.status !== 'end') {
      await this.client.quit();
    }
  }
}
