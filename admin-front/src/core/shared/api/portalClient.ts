import { assertApiSuccess, fetchWithNetworkFallback, parseApiErrorMessage } from './parseApiError';

/**
 * Базовый префикс портальных модулей, у которых пока нет Go-сервиса.
 * Когда появится реальный backend — достаточно поменять `PORTAL_API_PREFIX`
 * на `/api/v1` и удалить route-handlers из `src/app/portal-api`.
 */
export const PORTAL_API_PREFIX =
  process.env.NEXT_PUBLIC_PORTAL_API_PREFIX ?? '/portal-api';

export class PortalApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'PortalApiError';
    this.status = status;
  }
}

type PortalRequestOptions = {
  body?: unknown;
  fallback: string;
  method?: 'DELETE' | 'GET' | 'PATCH' | 'POST' | 'PUT';
  path: string;
  signal?: AbortSignal;
};

export const portalRequest = async <T>({
  body,
  fallback,
  method = 'GET',
  path,
  signal,
}: PortalRequestOptions): Promise<T> => {
  const response = await fetchWithNetworkFallback(
    `${PORTAL_API_PREFIX}${path}`,
    {
      body: body === undefined ? undefined : JSON.stringify(body),
      headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
      method,
      signal,
    },
    fallback,
  );

  if (!response.ok) {
    throw new PortalApiError(response.status, await parseApiErrorMessage(response, fallback));
  }

  const data: unknown = await response.json();
  const record = assertApiSuccess(data, fallback);

  return record.data as T;
};
