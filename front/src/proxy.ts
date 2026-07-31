import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

import { USER_ID_COOKIE } from '@/core/entities/session/lib/constants';
import { AppPath } from '@/core/shared/router/paths';

export const config = {
  matcher: ['/cabinet', '/cabinet/:path*', '/checkout'],
};

export const proxy = (request: NextRequest): NextResponse => {
  const { pathname } = request.nextUrl;

  const userId = request.cookies.get(USER_ID_COOKIE)?.value;

  if (!userId) {
    const loginUrl = new URL(AppPath.Login, request.url);

    loginUrl.searchParams.set('next', pathname);

    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
};
