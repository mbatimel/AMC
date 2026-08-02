/**
 * Безопасный внутренний путь из `?next=` (только relative, без protocol-relative).
 * Защита от open redirect после логина.
 */
export const resolveSafeNextPath = (next: null | string, fallback = '/'): string => {
  if (next && next.startsWith('/') && !next.startsWith('//')) {
    return next;
  }

  return fallback;
};
