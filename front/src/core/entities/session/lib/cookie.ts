import { USER_ID_COOKIE } from './constants';

const COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 30;

export const readUserIdCookie = (): null | string => {
  if (typeof document === 'undefined') {
    return null;
  }

  const match = document.cookie.match(new RegExp(`(?:^|; )${USER_ID_COOKIE}=([^;]*)`));

  if (!match?.[1]) {
    return null;
  }

  return decodeURIComponent(match[1]);
};

export const writeUserIdCookie = (userId: string): void => {
  document.cookie = `${USER_ID_COOKIE}=${encodeURIComponent(userId)}; Path=/; Max-Age=${COOKIE_MAX_AGE_SECONDS}; SameSite=Lax`;
};

export const clearUserIdCookie = (): void => {
  document.cookie = `${USER_ID_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
};
