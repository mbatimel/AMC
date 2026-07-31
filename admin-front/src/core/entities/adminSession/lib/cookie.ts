import { ADMIN_USER_ID_COOKIE } from './constants';

const COOKIE_MAX_AGE_SECONDS = 60 * 60 * 12;

export const readAdminUserIdCookie = (): null | string => {
  if (typeof document === 'undefined') {
    return null;
  }

  const match = document.cookie.match(new RegExp(`(?:^|; )${ADMIN_USER_ID_COOKIE}=([^;]*)`));

  if (!match?.[1]) {
    return null;
  }

  return decodeURIComponent(match[1]);
};

export const writeAdminUserIdCookie = (userId: string): void => {
  document.cookie = `${ADMIN_USER_ID_COOKIE}=${encodeURIComponent(userId)}; Path=/; Max-Age=${COOKIE_MAX_AGE_SECONDS}; SameSite=Lax`;
};

export const clearAdminUserIdCookie = (): void => {
  document.cookie = `${ADMIN_USER_ID_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
};
