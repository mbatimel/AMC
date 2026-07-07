import type { ApiEnvelope } from './types';

import { ApiError } from './errors';

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return typeof value === 'object' && value !== null;
};

const isEnvelope = (value: unknown): value is ApiEnvelope<unknown> => {
  return isRecord(value) && 'data' in value && 'error' in value && typeof value.error === 'boolean';
};

export const parseApiResponse = <T>(body: unknown, status: number): T => {
  if (!isEnvelope(body)) {
    return body as T;
  }

  if (body.error) {
    throw ApiError.fromEnvelope(body, status);
  }

  return body.data as T;
};
