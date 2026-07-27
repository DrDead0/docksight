import { SetMetadata } from '@nestjs/common';

export const IS_PUBLIC_KEY = 'docksight:public';

/**
 * Opts a route out of the globally registered JwtAuthGuard.
 *
 * Authentication is on by default for every route in the app, so a new
 * controller is protected the moment it is written. Only the three routes that
 * genuinely cannot require a token carry this decorator: first-run setup
 * (nobody exists yet) and login (the token is what you are asking for).
 */
export const Public = () => SetMetadata(IS_PUBLIC_KEY, true);
