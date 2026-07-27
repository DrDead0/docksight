import { Injectable, Logger, UnauthorizedException } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { UsersService, type SafeUser } from '../users/users.service';
import type { AuthUser, JwtPayload, LoginResult } from './auth.types';

/**
 * Credential verification and token issuing.
 *
 * It never touches Prisma directly — user lookup and password comparison stay
 * behind UsersService, so there is one implementation of "is this password
 * correct" in the codebase.
 */
@Injectable()
export class AuthService {
  private readonly logger = new Logger(AuthService.name);

  constructor(
    private readonly users: UsersService,
    private readonly jwt: JwtService,
  ) {}

  /**
   * Exchanges credentials for a signed access token.
   *
   * Both "no such account" and "wrong password" raise the same 401 with the
   * same message: distinguishing them turns the login route into an account
   * enumeration oracle.
   */
  async login(email: string, password: string): Promise<LoginResult> {
    const user = await this.users.validateCredentials(email, password);

    if (!user) {
      this.logger.warn(`Failed login attempt for ${redactEmail(email)}`);
      throw new UnauthorizedException('Invalid email or password');
    }

    const accessToken = await this.signToken(user);
    this.logger.log(`User logged in id=${user.id} role=${user.role}`);

    return { accessToken, user: toAuthUser(user) };
  }

  /**
   * Resolves the caller behind a validated token. The token carries only an id
   * and a role, so the profile is read back from the database — which also
   * means a deleted user stops being authenticated immediately, without
   * waiting for token expiry.
   */
  async getProfile(userId: string): Promise<AuthUser> {
    const user = await this.users.findById(userId);

    if (!user) {
      throw new UnauthorizedException('Account no longer exists');
    }

    return toAuthUser(user);
  }

  private signToken(user: SafeUser): Promise<string> {
    const payload: JwtPayload = { sub: user.id, role: user.role };
    return this.jwt.signAsync(payload);
  }
}

function toAuthUser(user: SafeUser): AuthUser {
  return { id: user.id, email: user.email, role: user.role };
}

/** Keeps full addresses out of the log file while staying useful for triage. */
function redactEmail(email: string): string {
  const [local, domain] = (email ?? '').split('@');
  if (!domain) {
    return '<invalid>';
  }
  return `${local.slice(0, 2)}***@${domain}`;
}
