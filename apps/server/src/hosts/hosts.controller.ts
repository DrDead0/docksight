import {
  Controller,
  Get,
  NotFoundException,
  Param,
} from '@nestjs/common';
import { ApiOkResponse, ApiOperation, ApiTags } from '@nestjs/swagger';
import { HostsService } from './hosts.service';

@ApiTags('hosts')
@Controller('hosts')
export class HostsController {
  constructor(private readonly hostsService: HostsService) {}

  @Get()
  @ApiOperation({ summary: 'List registered Docker hosts (agents)' })
  @ApiOkResponse({ description: 'Registered hosts' })
  listHosts() {
    return this.hostsService.listHosts();
  }

  @Get(':id/containers')
  @ApiOperation({ summary: 'List containers discovered on a host' })
  @ApiOkResponse({ description: 'Container inventory for the host' })
  async listContainers(@Param('id') id: string) {
    const result = await this.hostsService.listContainers(id);
    if (!result) {
      throw new NotFoundException(`Host not found: ${id}`);
    }
    return result;
  }
}
