import { SetMetadata } from '@nestjs/common';
import { UserRole } from '../../generated/prisma/client';

export const ROLES_KEY = 'docksight:roles';

/**
 * Declares which roles may call a route.
 *
 *   @Post(':id/restart')
 *   @Roles('ADMIN')
 *   restartContainer() {}
 *
 * Only meaningful alongside JwtAuthGuard + RolesGuard — the metadata is inert
 * on its own, which is why the guards are applied at the controller level.
 */
export const Roles = (...roles: UserRole[]) => SetMetadata(ROLES_KEY, roles);
