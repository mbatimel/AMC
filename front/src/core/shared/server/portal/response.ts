import { NextResponse } from 'next/server';

/** Ответы повторяют контракт Go-сервисов: `{ error, errorText, data }`. */
export const apiOk = <T>(data: T): NextResponse =>
  NextResponse.json({ data, error: false, errorText: '' });

export const apiFail = (status: number, errorText: string): NextResponse =>
  NextResponse.json({ data: null, error: true, errorText }, { status });

export const readJsonBody = async <T>(request: Request): Promise<null | T> => {
  try {
    return (await request.json()) as T;
  } catch {
    return null;
  }
};
