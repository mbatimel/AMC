import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

import { AUTH_CALLBACK_SEARCH_PARAM, AUTH_SESSION_COOKIE } from '@/core/entities/auth';
import { AppPath, isPublicPath } from '@/core/shared/router';

const AUTH_PATH: string = AppPath.Auth;

export const middleware = (request: NextRequest): NextResponse => {
  const { pathname } = request.nextUrl;
  const hasSession = Boolean(request.cookies.get(AUTH_SESSION_COOKIE)?.value);

  if (isPublicPath(pathname)) {
    if (hasSession && pathname === AUTH_PATH) {
      return NextResponse.redirect(new URL(AppPath.Home, request.url));
    }

    return NextResponse.next();
  }

  if (!hasSession) {
    const authUrl = new URL(AppPath.Auth, request.url);
    authUrl.searchParams.set(AUTH_CALLBACK_SEARCH_PARAM, pathname);

    return NextResponse.redirect(authUrl);
  }

  return NextResponse.next();
};

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)'],
};
