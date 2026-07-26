import {
  BadRequestException,
  Body,
  Controller,
  GatewayTimeoutException,
  HttpException,
  MessageEvent,
  NotFoundException,
  Param,
  Post,
  Query,
  Sse,
  ServiceUnavailableException,
} from '@nestjs/common';
import {
  ApiBody,
  ApiOkResponse,
  ApiOperation,
  ApiProperty,
  ApiTags,
} from '@nestjs/swagger';
import { IsNotEmpty, IsString } from 'class-validator';
import { Observable, finalize } from 'rxjs';
import type { LogsChunkPayload } from '@docksight/protocol';
import { ContainersService } from './containers.service';

class ContainerActionBodyDto {
  @ApiProperty({ description: 'Host (agent) id that owns the container' })
  @IsString()
  @IsNotEmpty()
  hostId!: string;
}

@ApiTags('containers')
@Controller('containers')
export class ContainersController {
  constructor(private readonly containersService: ContainersService) {}

  @Post(':id/start')
  @ApiOperation({ summary: 'Start a container on a host' })
  @ApiBody({ type: ContainerActionBodyDto })
  @ApiOkResponse({ description: 'Lifecycle command result' })
  start(
    @Param('id') containerId: string,
    @Body() body: ContainerActionBodyDto,
  ) {
    return this.runAction(containerId, body.hostId, 'start');
  }

  @Post(':id/stop')
  @ApiOperation({ summary: 'Stop a container on a host' })
  @ApiBody({ type: ContainerActionBodyDto })
  @ApiOkResponse({ description: 'Lifecycle command result' })
  stop(
    @Param('id') containerId: string,
    @Body() body: ContainerActionBodyDto,
  ) {
    return this.runAction(containerId, body.hostId, 'stop');
  }

  @Post(':id/restart')
  @ApiOperation({ summary: 'Restart a container on a host' })
  @ApiBody({ type: ContainerActionBodyDto })
  @ApiOkResponse({ description: 'Lifecycle command result' })
  restart(
    @Param('id') containerId: string,
    @Body() body: ContainerActionBodyDto,
  ) {
    return this.runAction(containerId, body.hostId, 'restart');
  }

  @Sse(':id/logs')
  @ApiOperation({
    summary: 'Stream container logs (SSE)',
    description:
      'Opens a live log stream via the connected agent. Query: hostId (required), tail, follow.',
  })
  streamLogs(
    @Param('id') containerId: string,
    @Query('hostId') hostId: string,
    @Query('tail') tailRaw?: string,
    @Query('follow') followRaw?: string,
  ): Observable<MessageEvent> {
    if (!hostId?.trim()) {
      throw new BadRequestException('hostId query parameter is required');
    }

    const tail = Number.parseInt(tailRaw ?? '100', 10);
    const follow = followRaw !== 'false';

    return new Observable<MessageEvent>((subscriber) => {
      let requestId: string | null = null;
      let closed = false;

      const onChunk = (payload: LogsChunkPayload) => {
        if (closed) {
          return;
        }
        subscriber.next({
          type: 'logs.chunk',
          data: payload,
        } as MessageEvent);
      };

      void this.containersService
        .startLogStream({
          hostId,
          containerId,
          tail: Number.isFinite(tail) && tail > 0 ? tail : 100,
          follow,
          onChunk,
        })
        .then((id) => {
          requestId = id;
          subscriber.next({
            type: 'logs.subscribed',
            data: { requestId: id, containerId },
          } as MessageEvent);
        })
        .catch((error: unknown) => {
          const message =
            error instanceof Error ? error.message : 'Failed to start log stream';
          subscriber.error(this.mapStreamError(message));
        });

      return () => {
        closed = true;
        if (requestId) {
          this.containersService.stopLogStream(requestId);
        }
      };
    }).pipe(
      finalize(() => {
        // teardown also handled by Observable unsubscribe return
      }),
    );
  }

  private mapStreamError(message: string): HttpException {
    if (message.includes('not found')) {
      return new NotFoundException(message);
    }
    if (message.includes('not connected') || message.includes('offline')) {
      return new ServiceUnavailableException(message);
    }
    return new BadRequestException(message);
  }

  private async runAction(
    containerId: string,
    hostId: string,
    action: 'start' | 'stop' | 'restart',
  ) {
    try {
      return await this.containersService.runLifecycleAction(
        hostId,
        containerId,
        action,
      );
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }
      const message =
        error instanceof Error ? error.message : 'Container action failed';

      if (message.includes('not found')) {
        throw new NotFoundException(message);
      }
      if (message.includes('not connected') || message.includes('offline')) {
        throw new ServiceUnavailableException(message);
      }
      if (message.includes('timed out')) {
        throw new GatewayTimeoutException(message);
      }
      if (message.includes('required')) {
        throw new BadRequestException(message);
      }
      throw new BadRequestException(message);
    }
  }
}
