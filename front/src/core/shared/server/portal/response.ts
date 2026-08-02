import { NextResponse } from 'next/server';

/**
 * Состояние portal-api меняется админкой в рантайме (без пересборки), поэтому
 * ответы не должны оседать ни в браузере, ни в промежуточных прокси/CDN —
 * иначе часть пользователей продолжит видеть версию контента до правки.
 */
const NO_STORE_HEADERS = { 'Cache-Control': 'no-store, no-cache, must-revalidate' };

/** Ответы повторяют контракт Go-сервисов: `{ error, errorText, data }`. */
export const apiOk = <T>(data: T): NextResponse =>
  NextResponse.json({ data, error: false, errorText: '' }, { headers: NO_STORE_HEADERS });

export const apiFail = (status: number, errorText: string): NextResponse =>
  NextResponse.json({ data: null, error: true, errorText }, { headers: NO_STORE_HEADERS, status });

export const readJsonBody = async <T>(request: Request): Promise<null | T> => {
  try {
    return (await request.json()) as T;
  } catch {
    return null;
  }
};
