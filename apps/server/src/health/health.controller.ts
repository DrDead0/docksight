import {
  Controller,
  Get,
  ServiceUnavailableException,
} from '@nestjs/common';
import {
  ApiOkResponse,
  ApiOperation,
  ApiServiceUnavailableResponse,
  ApiTags,
} from '@nestjs/swagger';
import { Public } from '../auth/public.decorator';
import { PrismaService } from '../common/database/prisma.service';

export type HealthResponse = {
  status: 'ok';
  database: 'up';
};

/**
 * Unauthenticated liveness/readiness probe. Compose and `docksight install`
 * treat a service without a healthcheck as ready as soon as the process is
 * running; this endpoint is what those checks actually hit.
 *
 * Mounted at `/health` (excluded from the global `api` prefix) so probes do
 * not need to know about the API mount path.
 */
@ApiTags('health')
@Controller('health')
@Public()
export class HealthController {
  constructor(private readonly prisma: PrismaService) {}

  @Get()
  @ApiOperation({
    summary: 'Process and database readiness',
    description:
      'Returns 200 when the API can reach PostgreSQL. Used by the compose healthcheck.',
  })
  @ApiOkResponse({ description: 'API and database are reachable' })
  @ApiServiceUnavailableResponse({
    description: 'Database is unreachable',
  })
  async check(): Promise<HealthResponse> {
    try {
      await this.prisma.$queryRaw`SELECT 1`;
    } catch {
      throw new ServiceUnavailableException({
        status: 'error',
        database: 'down',
      });
    }

    return { status: 'ok', database: 'up' };
  }
}
