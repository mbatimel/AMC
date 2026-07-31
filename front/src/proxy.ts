import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

import { ADMIN_USER_ID_COOKIE } from '@/core/entities/adminSession/lib/constants';
import { USER_ID_COOKIE } from '@/core/entities/session/lib/constants';
import { AppPath } from '@/core/shared/router/paths';

export const config = {
  matcher: ['/admin', '/admin/:path*', '/cabinet', '/cabinet/:path*', '/checkout'],
};

export const proxy = (request: NextRequest): NextResponse => {
  const { pathname } = request.nextUrl;

  if (pathname.startsWith(AppPath.Admin)) {
    if (pathname === AppPath.AdminLogin) {
      return NextResponse.next();
    }

    const adminUserId = request.cookies.get(ADMIN_USER_ID_COOKIE)?.value;

    if (!adminUserId) {
      return NextResponse.redirect(new URL(AppPath.AdminLogin, request.url));
    }

    return NextResponse.next();
  }

  const userId = request.cookies.get(USER_ID_COOKIE)?.value;

  if (!userId) {
    const loginUrl = new URL(AppPath.Login, request.url);

    loginUrl.searchParams.set('next', pathname);

    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
};
