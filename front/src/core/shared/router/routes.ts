import { AppPath } from './paths';

export const PUBLIC_PATHS: readonly string[] = [AppPath.Auth];

export const isPublicPath = (pathname: string): boolean => {
  return PUBLIC_PATHS.some(
    (path) => pathname === path || pathname.startsWith(`${path}/`),
  );
};
