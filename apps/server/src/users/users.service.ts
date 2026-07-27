import {
  BadRequestException,
  ConflictException,
  Injectable,
  Logger,
} from '@nestjs/common';
import { Prisma, User, UserRole } from '../../generated/prisma/client';
import { PrismaService } from '../common/database/prisma.service';
import { hashPassword, verifyPassword } from './password.util';

export type RegisterUserInput = {
  email: string;
  password: string;
  /** Optional display handle. First-run setup only collects email + password. */
  username?: string | null;
  /** Defaults to VIEWER. Only SetupModule is allowed to pass ADMIN. */
  role?: UserRole;
};

/** A user record with the password hash removed — safe to return over HTTP. */
export type SafeUser = Omit<User, 'passwordHash'>;

const MIN_USERNAME_LENGTH = 3;
const MIN_PASSWORD_LENGTH = 8;

/**
 * The only module that reads or writes the users table. Everything above it
 * (AuthService, SetupService) goes through this service, so password hashing
 * and the "never leak passwordHash" rule live in exactly one place.
 */
@Injectable()
export class UsersService {
  private readonly logger = new Logger(UsersService.name);

  constructor(private readonly prisma: PrismaService) {}

  /**
   * Creates a user with a bcrypt-hashed password.
   * Throws ConflictException when the email or username is already taken.
   */
  async register(input: RegisterUserInput): Promise<SafeUser> {
    const email = normalizeEmail(input.email);
    const password = input.password ?? '';
    const username = input.username?.trim() || null;

    if (!email) {
      throw new BadRequestException('Email is required');
    }
    if (password.length < MIN_PASSWORD_LENGTH) {
      throw new BadRequestException(
        `Password must be at least ${MIN_PASSWORD_LENGTH} characters`,
      );
    }
    if (username !== null && username.length < MIN_USERNAME_LENGTH) {
      throw new BadRequestException(
        `Username must be at least ${MIN_USERNAME_LENGTH} characters`,
      );
    }

    // Checked up front so the caller learns *which* field collided; the P2002
    // catch below is the backstop for two registrations racing each other.
    const [emailTaken, usernameTaken] = await Promise.all([
      this.prisma.user.findUnique({ where: { email }, select: { id: true } }),
      username
        ? this.prisma.user.findUnique({
            where: { username },
            select: { id: true },
          })
        : Promise.resolve(null),
    ]);
    if (emailTaken) {
      throw new ConflictException('Email is already registered');
    }
    if (usernameTaken) {
      throw new ConflictException('Username is already taken');
    }

    const passwordHash = await hashPassword(password);

    try {
      const user = await this.prisma.user.create({
        data: {
          email,
          username,
          passwordHash,
          role: input.role ?? UserRole.VIEWER,
        },
      });

      this.logger.log(`User created id=${user.id} role=${user.role}`);
      return toSafeUser(user);
    } catch (error) {
      if (
        error instanceof Prisma.PrismaClientKnownRequestError &&
        error.code === 'P2002'
      ) {
        throw new ConflictException(
          'That email or username is already registered',
        );
      }
      throw error;
    }
  }

  /** Looks a user up by email. Email matching is case-insensitive. */
  async findByEmail(email: string): Promise<SafeUser | null> {
    const user = await this.findByEmailWithSecret(email);
    return user ? toSafeUser(user) : null;
  }

  async findById(id: string): Promise<SafeUser | null> {
    if (!id?.trim()) {
      return null;
    }

    const user = await this.prisma.user.findUnique({ where: { id } });
    return user ? toSafeUser(user) : null;
  }

  async findByUsername(username: string): Promise<SafeUser | null> {
    const value = username?.trim();
    if (!value) {
      return null;
    }

    const user = await this.prisma.user.findUnique({
      where: { username: value },
    });
    return user ? toSafeUser(user) : null;
  }

  /**
   * Full record including `passwordHash` — for the auth flow only, never for a
   * response body. Public reads should use `findByEmail`.
   */
  findByEmailWithSecret(email: string): Promise<User | null> {
    const value = normalizeEmail(email);
    if (!value) {
      return Promise.resolve(null);
    }

    return this.prisma.user.findUnique({ where: { email: value } });
  }

  /**
   * Verifies a login pair. Returns the user on success, null otherwise —
   * a missing user and a wrong password are deliberately indistinguishable.
   */
  async validateCredentials(
    email: string,
    password: string,
  ): Promise<SafeUser | null> {
    const user = await this.findByEmailWithSecret(email);
    if (!user) {
      // Hash anyway so a missing account and a wrong password take the same
      // time — otherwise response latency reveals which emails are registered.
      await verifyPassword(password ?? '', DUMMY_HASH);
      return null;
    }

    const valid = await verifyPassword(password ?? '', user.passwordHash);
    return valid ? toSafeUser(user) : null;
  }

  /** Whether any user exists — used by first-run setup to gate registration. */
  async hasAnyUser(): Promise<boolean> {
    const count = await this.prisma.user.count();
    return count > 0;
  }

  countUsers(): Promise<number> {
    return this.prisma.user.count();
  }
}

/**
 * A real bcrypt hash of a random value, used only to burn the same CPU time as
 * a genuine comparison when the account does not exist.
 */
const DUMMY_HASH =
  '$2b$12$C6UzMDM.H6dfI/f/IKcEe.7rW9RGmYVRPzO0hqcVUCg7SBnMEfWLu';

function normalizeEmail(email: string | undefined): string {
  return (email ?? '').trim().toLowerCase();
}

/**
 * Explicit allow-list rather than `delete user.passwordHash`: if a secret
 * column is added to the model later, it stays out of responses by default
 * instead of leaking until someone remembers to strip it.
 */
function toSafeUser(user: User): SafeUser {
  return {
    id: user.id,
    username: user.username,
    email: user.email,
    role: user.role,
    createdAt: user.createdAt,
    updatedAt: user.updatedAt,
  };
}
