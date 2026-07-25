import {
  BadRequestException,
  Body,
  Controller,
  GatewayTimeoutException,
  HttpException,
  NotFoundException,
  Param,
  Post,
  ServiceUnavailableException,
} from '@nestjs/common';
import { ApiBody, ApiOkResponse, ApiOperation, ApiProperty, ApiTags } from '@nestjs/swagger';
import { IsNotEmpty, IsString } from 'class-validator';
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
