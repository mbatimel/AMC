import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

import { ADMIN_USER_ID_COOKIE } from '@/core/entities/adminSession/lib/constants';
import { AppPath } from '@/core/shared/router/paths';
import { resolveSafeNextPath } from '@/core/shared/router/resolveSafeNextPath';

export const config = {
  matcher: ['/', '/((?!_next|api|portal-api|favicon.ico).*)'],
};

export const proxy = (request: NextRequest): NextResponse => {
  const { pathname } = request.nextUrl;
  const adminUserId = request.cookies.get(ADMIN_USER_ID_COOKIE)?.value;
  const isLoginRoute = pathname === (AppPath.Login as string);

  if (isLoginRoute) {
    if (adminUserId) {
      const target = resolveSafeNextPath(request.nextUrl.searchParams.get('next'));

      return NextResponse.redirect(new URL(target, request.url));
    }

    return NextResponse.next();
  }

  if (!adminUserId) {
    const loginUrl = new URL(AppPath.Login, request.url);

    loginUrl.searchParams.set('next', pathname);

    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
};
