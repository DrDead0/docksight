import { ConflictException, Injectable, Logger } from '@nestjs/common';
import { UserRole } from '../../generated/prisma/client';
import { UsersService } from '../users/users.service';
import type { AuthUser } from '../auth/auth.types';

export type SetupStatus = {
  setupRequired: boolean;
};

/**
 * First-run bootstrap.
 *
 * This is the one route that creates a user without being authenticated, so
 * the "users table is empty" check is the entire security control. It lives
 * here rather than in the controller because it is a business rule, and it is
 * re-checked immediately before the write.
 */
@Injectable()
export class SetupService {
  private readonly logger = new Logger(SetupService.name);

  constructor(private readonly users: UsersService) {}

  /** Whether DockSight still needs its first administrator. */
  async getStatus(): Promise<SetupStatus> {
    const hasUsers = await this.users.hasAnyUser();
    return { setupRequired: !hasUsers };
  }

  /**
   * Creates the first administrator. Refuses once any user exists — otherwise
   * this endpoint would be a permanent unauthenticated way to mint an ADMIN.
   *
   * The role is hardcoded, never taken from the request body: accepting a role
   * from an anonymous caller is how privilege escalation bugs start.
   */
  async createFirstAdmin(email: string, password: string): Promise<AuthUser> {
    if (await this.users.hasAnyUser()) {
      throw new ConflictException('DockSight has already been set up');
    }

    const user = await this.users.register({
      email,
      password,
      role: UserRole.ADMIN,
    });

    this.logger.log(`First administrator created id=${user.id}`);

    return { id: user.id, email: user.email, role: user.role };
  }
}
