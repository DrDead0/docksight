import * as bcrypt from 'bcrypt';

/**
 * Cost factor. 12 is the current sensible default: ~250ms per hash on typical
 * hardware, which is slow enough to make offline cracking expensive and fast
 * enough for an interactive login.
 */
const BCRYPT_ROUNDS = 12;

/** Hashes a plaintext password. The salt is generated per call by bcrypt. */
export function hashPassword(password: string): Promise<string> {
  return bcrypt.hash(password, BCRYPT_ROUNDS);
}

/**
 * Constant-time comparison of a candidate against a stored bcrypt hash.
 * Returns false (never throws) for a malformed or unknown-format hash, so a
 * corrupted row cannot crash the login route.
 */
export async function verifyPassword(
  password: string,
  hash: string,
): Promise<boolean> {
  if (!password || !hash) {
    return false;
  }

  try {
    return await bcrypt.compare(password, hash);
  } catch {
    return false;
  }
}
