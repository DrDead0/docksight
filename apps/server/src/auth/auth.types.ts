import { UserRole } from '../../generated/prisma/client';

/**
 * Signed JWT body. Kept to the identifier and the role only — anything else
 * (email, name) would be readable by anyone holding the token, since a JWT is
 * signed but not encrypted.
 */
export type JwtPayload = {
  sub: string;
  role: UserRole;
};

/** What JwtAuthGuard attaches to `request.user`. */
export type AuthenticatedUser = {
  id: string;
  role: UserRole;
};

/** The user shape returned to clients. Never contains the password hash. */
export type AuthUser = {
  id: string;
  email: string;
  role: UserRole;
};

export type LoginResult = {
  accessToken: string;
  user: AuthUser;
};

/** Express request once JwtAuthGuard has run. */
export type RequestWithUser = {
  user?: AuthenticatedUser;
  headers: Record<string, string | string[] | undefined>;
  url?: string;
  query?: Record<string, string | string[] | undefined>;
};
