export const parseApiErrorMessage = async (
  response: Response,
  fallback: string,
): Promise<string> => {
  try {
    const data: unknown = await response.json();

    if (typeof data === 'object' && data !== null) {
      const record = data as Record<string, unknown>;
      const errorText = record.errorText ?? record.ErrorText;

      if (typeof errorText === 'string' && errorText.length > 0) {
        return errorText;
      }
    }
  } catch {
    // ignore JSON parse errors
  }

  if (response.status === 502 || response.status === 503 || response.status === 504) {
    return 'Сервис временно недоступен. Попробуйте позже';
  }

  return fallback;
};

/** Браузерные сетевые ошибки (`Failed to fetch` и т.п.) → русский fallback. */
export const toDisplayErrorMessage = (error: unknown, fallback: string): string => {
  if (!(error instanceof Error) || error.message.trim().length === 0) {
    return fallback;
  }

  const message = error.message.trim();
  const normalized = message.toLowerCase();

  if (
    normalized === 'failed to fetch' ||
    normalized === 'load failed' ||
    normalized === 'network request failed' ||
    normalized.includes('networkerror when attempting to fetch') ||
    normalized.includes('fetch failed')
  ) {
    return fallback;
  }

  return message;
};

export const fetchWithNetworkFallback = async (
  input: RequestInfo | URL,
  init: RequestInit | undefined,
  networkFallback: string,
): Promise<Response> => {
  try {
    return await fetch(input, init);
  } catch (error) {
    throw new Error(toDisplayErrorMessage(error, networkFallback), { cause: error });
  }
};

export const assertApiSuccess = (data: unknown, fallback: string): Record<string, unknown> => {
  if (typeof data !== 'object' || data === null) {
    throw new Error(fallback);
  }

  const record = data as Record<string, unknown>;

  if (record.error === true) {
    const errorText = record.errorText ?? record.ErrorText;

    throw new Error(typeof errorText === 'string' && errorText.length > 0 ? errorText : fallback);
  }

  return record;
};

export const omitEmptyParams = (
  params: Record<string, boolean | number | string | undefined>,
): string => {
  const search = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === '') {
      return;
    }

    search.set(key, String(value));
  });

  const query = search.toString();

  return query.length > 0 ? `?${query}` : '';
};
