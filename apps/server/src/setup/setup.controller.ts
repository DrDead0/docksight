import { Body, Controller, Get, Post } from '@nestjs/common';
import {
  ApiConflictResponse,
  ApiCreatedResponse,
  ApiOkResponse,
  ApiOperation,
  ApiTags,
} from '@nestjs/swagger';
import { Public } from '../auth/public.decorator';
import { CreateAdminDto } from './dto/create-admin.dto';
import { SetupService, type SetupStatus } from './setup.service';
import type { AuthUser } from '../auth/auth.types';

/**
 * Both routes are intentionally unauthenticated — there is no one to
 * authenticate as before the first user exists. `create-admin` protects itself
 * by refusing to run once the users table is non-empty.
 */
@ApiTags('setup')
@Controller('setup')
// Both routes opt out of the global JwtAuthGuard: there is no one to
// authenticate as before the first user exists. `create-admin` protects itself
// by refusing to run once the users table is non-empty.
@Public()
export class SetupController {
  constructor(private readonly setupService: SetupService) {}

  @Get('status')
  @ApiOperation({ summary: 'Whether DockSight still needs its first admin' })
  @ApiOkResponse({ schema: { example: { setupRequired: true } } })
  getStatus(): Promise<SetupStatus> {
    return this.setupService.getStatus();
  }

  @Post('create-admin')
  @ApiOperation({ summary: 'Create the first administrator account' })
  @ApiCreatedResponse({ description: 'The created administrator' })
  @ApiConflictResponse({ description: 'DockSight has already been set up' })
  createAdmin(@Body() body: CreateAdminDto): Promise<AuthUser> {
    return this.setupService.createFirstAdmin(body.email, body.password);
  }
}
