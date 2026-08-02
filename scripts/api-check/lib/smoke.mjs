import { normalizePath } from './inventory.mjs';

const PLACEHOLDER_UUID = '00000000-0000-4000-8000-000000000001';

/** Пути, где nginx требует trailing slash (как во фронте). */
const TRAILING_SLASH_PATHS = new Set(['/api/v1/orders']);

const fillPathParams = (endpointPath) => {
  let value = endpointPath.replace(/\{[^}]+\}/g, PLACEHOLDER_UUID);

  if (TRAILING_SLASH_PATHS.has(value)) {
    value = `${value}/`;
  }

  return value;
};

const classifyResult = ({ envelopeError, hasPathParams, ok, status, transport }) => {
  if (status === 0) {
    return 'network';
  }

  if (status === 501) {
    return 'fail';
  }

  // Плейсхолдер UUID часто даёт 404 — ручка жива.
  if (status === 404 && hasPathParams) {
    return 'ok';
  }

  if (status === 404) {
    return 'fail';
  }

  if (status === 301 || status === 302 || status === 307 || status === 308) {
    return 'fail';
  }

  if (status === 401 || status === 403) {
    return transport === 'api' ? 'auth' : 'fail';
  }

  if (status >= 500) {
    // Без X-User-Id orders часто отвечает 500 — это сигнал «нужна сессия», не «ручки нет».
    return 'auth';
  }

  if (envelopeError) {
    return 'envelope';
  }

  if (ok || status === 400 || status === 409 || status === 422) {
    return 'ok';
  }

  return 'fail';
};

const DEFAULT_SKIP = {
  /** Не гоняем по умолчанию — меняют данные. */
  destructiveMethods: new Set(['DELETE']),
};

const parseEnvelope = async (response) => {
  const contentType = response.headers.get('content-type') ?? '';

  if (!contentType.includes('application/json')) {
    return { envelopeError: false, errorText: null };
  }

  try {
    const data = await response.json();

    if (typeof data === 'object' && data !== null && data.error === true) {
      return {
        envelopeError: true,
        errorText: typeof data.errorText === 'string' ? data.errorText : 'error=true',
      };
    }
  } catch {
    return { envelopeError: false, errorText: null };
  }

  return { envelopeError: false, errorText: null };
};

const requestOnce = async ({ baseUrl, headers, method, endpointPath, timeoutMs }) => {
  const url = new URL(fillPathParams(endpointPath), baseUrl.endsWith('/') ? baseUrl : `${baseUrl}/`);
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      headers,
      method,
      redirect: 'manual',
      signal: controller.signal,
    });

    const envelope = await parseEnvelope(response);

    return {
      ...envelope,
      ok: response.ok,
      status: response.status,
    };
  } catch (error) {
    return {
      envelopeError: false,
      errorText: error instanceof Error ? error.message : 'network error',
      ok: false,
      status: 0,
    };
  } finally {
    clearTimeout(timer);
  }
};

const loginUser = async (apiBase, email, password) => {
  const response = await fetch(new URL('/api/v1/auth/login', apiBase), {
    body: JSON.stringify({ email, password }),
    headers: { 'Content-Type': 'application/json' },
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error(`user login failed: HTTP ${response.status}`);
  }

  const data = await response.json();
  const userID = data?.data?.userID ?? data?.userID;

  if (typeof userID !== 'string' || userID.length === 0) {
    throw new Error('user login failed: no userID in response');
  }

  return userID;
};

const loginAdmin = async (apiBase, email, password) => {
  const query = new URLSearchParams({ email, password });
  const response = await fetch(new URL(`/api/v1/admin/auth/login?${query}`, apiBase), {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error(`admin login failed: HTTP ${response.status}`);
  }

  const data = await response.json();
  const userID = data?.data?.userID ?? data?.userID;

  if (typeof userID !== 'string' || userID.length === 0) {
    throw new Error('admin login failed: no userID in response');
  }

  return userID;
};

export const runSmoke = async ({
  apiBase,
  endpoints,
  includeDestructive = false,
  includePortal = false,
  includeWrite = false,
  portalBase,
  timeoutMs = 15000,
  userEmail,
  userPassword,
  adminEmail,
  adminPassword,
}) => {
  let userId = null;
  let adminId = null;

  if (userEmail && userPassword) {
    userId = await loginUser(apiBase, userEmail, userPassword);
  }

  if (adminEmail && adminPassword) {
    adminId = await loginAdmin(apiBase, adminEmail, adminPassword);
  }

  const results = [];

  for (const endpoint of endpoints) {
    if (endpoint.transport === 'portal' && !includePortal) {
      results.push({
        ...endpoint,
        detail: 'skipped (portal; use --portal)',
        status: null,
        verdict: 'skip',
      });
      continue;
    }

    if (DEFAULT_SKIP.destructiveMethods.has(endpoint.method) && !includeDestructive) {
      results.push({
        ...endpoint,
        detail: 'skipped (destructive; use --destructive)',
        status: null,
        verdict: 'skip',
      });
      continue;
    }

    if (
      !includeWrite &&
      endpoint.method !== 'GET' &&
      !(endpoint.method === 'POST' && /\/auth\/login$/.test(endpoint.path))
    ) {
      results.push({
        ...endpoint,
        detail: 'skipped (non-GET; use --write)',
        status: null,
        verdict: 'skip',
      });
      continue;
    }

    const baseUrl =
      endpoint.transport === 'portal'
        ? portalBase || apiBase
        : apiBase;

    const headers = {};
    const needsAdmin = endpoint.path.includes('/api/v1/admin/');
    const needsUser =
      endpoint.path.includes('/orders') ||
      endpoint.path.includes('/users/profile') ||
      endpoint.path.includes('/change-password') ||
      endpoint.path.includes('/favorites');

    if (needsAdmin && adminId) {
      headers['X-User-Id'] = adminId;
    } else if (needsUser && userId) {
      headers['X-User-Id'] = userId;
    }

    const response = await requestOnce({
      baseUrl,
      endpointPath: normalizePath(endpoint.path),
      headers,
      method: endpoint.method,
      timeoutMs,
    });

    const verdict = classifyResult({
      envelopeError: response.envelopeError,
      hasPathParams: /\{[^}]+\}/.test(endpoint.path),
      ok: response.ok,
      status: response.status,
      transport: endpoint.transport,
    });

    results.push({
      ...endpoint,
      detail: response.errorText,
      status: response.status,
      verdict,
    });
  }

  return { adminId, results, userId };
};
