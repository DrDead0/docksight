import {
  CanActivate,
  ExecutionContext,
  Injectable,
  UnauthorizedException,
} from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { JwtService } from '@nestjs/jwt';
import { IS_PUBLIC_KEY } from './public.decorator';
import type {
  AuthenticatedUser,
  JwtPayload,
  RequestWithUser,
} from './auth.types';

/**
 * Authentication guard, registered globally via APP_GUARD.
 *
 * Pulls the Bearer token off the Authorization header, verifies its signature
 * and expiry with the server secret, and attaches `{ id, role }` to the
 * request. Anything downstream can then trust `request.user` without re-doing
 * verification.
 *
 * Implemented directly on JwtService rather than passport-jwt: the whole job is
 * a dozen lines, and skipping passport keeps three dependencies out of the tree.
 */
@Injectable()
export class JwtAuthGuard implements CanActivate {
  constructor(
    private readonly jwt: JwtService,
    private readonly reflector: Reflector,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    // The agent WebSocket gateway authenticates agents, not users, and never
    // carries a user token. Only HTTP traffic is in scope here.
    if (context.getType() !== 'http') {
      return true;
    }

    const isPublic = this.reflector.getAllAndOverride<boolean>(IS_PUBLIC_KEY, [
      context.getHandler(),
      context.getClass(),
    ]);
    if (isPublic) {
      return true;
    }

    const request = context.switchToHttp().getRequest<RequestWithUser>();
    const token = extractToken(request);

    if (!token) {
      throw new UnauthorizedException('Missing authentication token');
    }

    let payload: JwtPayload;
    try {
      payload = await this.jwt.verifyAsync<JwtPayload>(token);
    } catch {
      // Covers a bad signature, a tampered payload and an expired token alike.
      // The client gets one message; the distinction is not its business.
      throw new UnauthorizedException('Invalid or expired token');
    }

    if (!payload?.sub || !payload?.role) {
      throw new UnauthorizedException('Malformed token payload');
    }

    const user: AuthenticatedUser = { id: payload.sub, role: payload.role };
    request.user = user;

    return true;
  }
}

function extractToken(request: RequestWithUser): string | null {
  return fromAuthorizationHeader(request) ?? fromEventSourceQuery(request);
}

function fromAuthorizationHeader(request: RequestWithUser): string | null {
  const header = request.headers?.authorization;
  const value = Array.isArray(header) ? header[0] : header;

  if (typeof value !== 'string') {
    return null;
  }

  const [scheme, token] = value.split(' ');
  if (scheme?.toLowerCase() !== 'bearer' || !token) {
    return null;
  }

  return token.trim();
}

/**
 * Fallback for `EventSource`, which cannot set request headers.
 *
 * A token in a query string is weaker than one in a header — it lands in
 * server access logs and browser history — so it is accepted only on the log
 * stream route, and the web client uses a fetch-based reader with a proper
 * Authorization header instead. This exists for curl and third-party clients.
 */
function fromEventSourceQuery(request: RequestWithUser): string | null {
  if (!request.url?.includes('/logs')) {
    return null;
  }

  const token = request.query?.access_token;
  return typeof token === 'string' && token.trim() ? token.trim() : null;
}
