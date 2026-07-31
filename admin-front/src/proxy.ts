import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

import { ADMIN_USER_ID_COOKIE } from '@/core/entities/adminSession/lib/constants';
import { AppPath } from '@/core/shared/router/paths';

export const config = {
  matcher: ['/', '/((?!_next|favicon.ico).*)'],
};

export const proxy = (request: NextRequest): NextResponse => {
  const { pathname } = request.nextUrl;

  if (pathname === AppPath.Login) {
    return NextResponse.next();
  }

  const adminUserId = request.cookies.get(ADMIN_USER_ID_COOKIE)?.value;

  if (!adminUserId) {
    return NextResponse.redirect(new URL(AppPath.Login, request.url));
  }

  return NextResponse.next();
};
