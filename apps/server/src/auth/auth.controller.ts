import {
  Body,
  Controller,
  Get,
  HttpCode,
  HttpStatus,
  Post,
} from '@nestjs/common';
import {
  ApiBearerAuth,
  ApiOkResponse,
  ApiOperation,
  ApiTags,
  ApiUnauthorizedResponse,
} from '@nestjs/swagger';
import { AuthService } from './auth.service';
import { CurrentUser } from './current-user.decorator';
import { LoginDto } from './dto/login.dto';
import { Public } from './public.decorator';
import type { AuthenticatedUser, AuthUser, LoginResult } from './auth.types';

/**
 * HTTP surface only: bind the request, hand off to AuthService, return the
 * result. No credential checks, no token handling, no database access here.
 */
@ApiTags('auth')
@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Post('login')
  // Public by necessity: a token is the thing being requested.
  @Public()
  // 200 rather than the POST default of 201 — logging in creates no resource.
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Exchange email and password for an access token' })
  @ApiOkResponse({ description: 'Access token and the signed-in user' })
  @ApiUnauthorizedResponse({ description: 'Invalid email or password' })
  login(@Body() body: LoginDto): Promise<LoginResult> {
    return this.authService.login(body.email, body.password);
  }

  @Get('me')
  // No @UseGuards needed — JwtAuthGuard is registered globally.
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Profile of the currently authenticated user' })
  @ApiOkResponse({ description: 'The authenticated user' })
  @ApiUnauthorizedResponse({ description: 'Missing, invalid or expired token' })
  me(@CurrentUser() user: AuthenticatedUser): Promise<AuthUser> {
    return this.authService.getProfile(user.id);
  }
}
