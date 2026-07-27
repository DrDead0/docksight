import { registerAs } from '@nestjs/config';
import type { JwtSignOptions } from '@nestjs/jwt';

export type AuthEnv = {
  jwtSecret: string;
  jwtExpiresIn: JwtSignOptions['expiresIn'];
};

const MIN_SECRET_LENGTH = 32;
const DEFAULT_EXPIRES_IN = '12h';

/** Seconds as a bare number, or a `ms` duration such as 30m / 12h / 7d. */
const DURATION_PATTERN = /^\d+(\.\d+)?\s*(ms|s|m|h|d|w|y)?$/i;

/**
 * JWT settings, sourced only from the environment.
 *
 * There is deliberately no fallback secret: a hardcoded default would let a
 * production deployment boot with a signing key that is public knowledge, and
 * anyone could then mint their own ADMIN token. Failing to start is the safe
 * outcome.
 */
export const authConfig = registerAs('auth', (): AuthEnv => {
  const jwtSecret = process.env.JWT_SECRET ?? '';

  if (!jwtSecret) {
    throw new Error(
      "JWT_SECRET is required. Generate one with: node -e \"console.log(require('crypto').randomBytes(48).toString('base64url'))\"",
    );
  }
  if (jwtSecret.length < MIN_SECRET_LENGTH) {
    throw new Error(
      `JWT_SECRET must be at least ${MIN_SECRET_LENGTH} characters (got ${jwtSecret.length}).`,
    );
  }

  // No refresh tokens in the MVP, so the access token has to live long enough
  // to be usable for a working session.
  const expiresIn = (process.env.JWT_EXPIRES_IN ?? DEFAULT_EXPIRES_IN).trim();

  if (!DURATION_PATTERN.test(expiresIn)) {
    throw new Error(
      `JWT_EXPIRES_IN must be a duration like "900", "30m" or "12h" (got "${expiresIn}").`,
    );
  }

  return {
    jwtSecret,
    // Safe after the format check above: jsonwebtoken accepts exactly these
    // shapes, but the type is a template literal that a plain env string
    // cannot satisfy on its own.
    jwtExpiresIn: expiresIn as JwtSignOptions['expiresIn'],
  };
});
