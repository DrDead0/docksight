import {
  CanActivate,
  ExecutionContext,
  ForbiddenException,
  Injectable,
} from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { UserRole } from '../../generated/prisma/client';
import { ROLES_KEY } from './roles.decorator';
import type { RequestWithUser } from './auth.types';

/**
 * Authorization guard. Runs after JwtAuthGuard and compares the caller's role
 * against whatever `@Roles(...)` declared on the handler (falling back to the
 * controller-level declaration).
 *
 * A route with no @Roles metadata is authenticated but not role-restricted.
 */
@Injectable()
export class RolesGuard implements CanActivate {
  constructor(private readonly reflector: Reflector) {}

  canActivate(context: ExecutionContext): boolean {
    if (context.getType() !== 'http') {
      return true;
    }

    const required = this.reflector.getAllAndOverride<UserRole[] | undefined>(
      ROLES_KEY,
      [context.getHandler(), context.getClass()],
    );

    if (!required || required.length === 0) {
      return true;
    }

    const request = context.switchToHttp().getRequest<RequestWithUser>();
    const user = request.user;

    if (!user) {
      // JwtAuthGuard should have run first; if it did not, fail closed.
      throw new ForbiddenException('Authentication required');
    }

    if (!required.includes(user.role)) {
      throw new ForbiddenException(
        `This action requires one of: ${required.join(', ')}`,
      );
    }

    return true;
  }
}
