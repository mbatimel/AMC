import type { AuditLogEntry } from '@/core/shared/server/portal/types';

import {
  assertApiSuccess,
  fetchWithNetworkFallback,
  omitEmptyParams,
  parseApiErrorMessage,
} from './parseApiError';
import { portalRequest } from './portalClient';

export type AdminAuditLogResult = {
  items: AdminAuditLogEntry[];
  total: number;
};

export type AdminAuditLogEntry = {
  action: string;
  actorLabel: string;
  createdAt: string;
  id: string;
};

export type AdminSessionResponse = {
  role: string;
  userID: string;
};

export class AdminApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'AdminApiError';
    this.status = status;
  }
}

const asString = (value: unknown, fallback = ''): string =>
  typeof value === 'string' ? value : fallback;

const adminHeaders = (userId: string): HeadersInit => ({ 'X-User-Id': userId });

const parseSession = (data: unknown, fallback: string): AdminSessionResponse => {
  const record = assertApiSuccess(data, fallback);
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new AdminApiError(500, fallback);
  }

  const source = payload as Record<string, unknown>;
  const userID = asString(source.userID);

  if (!userID) {
    throw new AdminApiError(500, fallback);
  }

  return { role: asString(source.role, 'admin'), userID };
};

/** Вход сотрудника портала. Бэк проверяет роль `admin` в access-сервисе. */
export const adminLoginRequest = async (credentials: {
  email: string;
  password: string;
}): Promise<AdminSessionResponse> => {
  const fallback = 'Не удалось выполнить вход';
  const query = omitEmptyParams({
    email: credentials.email,
    password: credentials.password,
  });
  const response = await fetchWithNetworkFallback(
    `/api/v1/admin/auth/login${query}`,
    { method: 'POST' },
    fallback,
  );

  if (!response.ok) {
    throw new AdminApiError(
      response.status,
      await parseApiErrorMessage(
        response,
        response.status === 401 || response.status === 403
          ? 'Неверный e-mail или пароль, либо нет прав администратора'
          : fallback,
      ),
    );
  }

  return parseSession(await response.json(), fallback);
};

export const adminSessionRequest = async (userId: string): Promise<AdminSessionResponse> => {
  const fallback = 'Сессия администратора недействительна';
  const response = await fetchWithNetworkFallback(
    '/api/v1/admin/auth/session',
    { headers: adminHeaders(userId) },
    fallback,
  );

  if (!response.ok) {
    throw new AdminApiError(response.status, await parseApiErrorMessage(response, fallback));
  }

  return parseSession(await response.json(), fallback);
};

export const adminLogoutRequest = async (userId: string): Promise<void> => {
  try {
    await fetch('/api/v1/admin/auth/logout', {
      headers: adminHeaders(userId),
      method: 'POST',
    });
  } catch {
    // выход всегда завершаем локально
  }
};

const parseAuditLog = (data: unknown, fallback: string): AdminAuditLogResult => {
  const record = assertApiSuccess(data, fallback);
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new AdminApiError(500, fallback);
  }

  const source = payload as Record<string, unknown>;
  const rawItems = Array.isArray(source.items) ? source.items : [];
  const items = rawItems.flatMap((item) => {
    if (typeof item !== 'object' || item === null) {
      return [];
    }

    const entry = item as Record<string, unknown>;

    return [
      {
        action: asString(entry.action),
        actorLabel: asString(entry.actorLabel, 'Админ портала'),
        createdAt: asString(entry.createdAt),
        id: asString(entry.id, asString(entry.createdAt)),
      },
    ];
  });

  return {
    items,
    total: typeof source.total === 'number' ? source.total : items.length,
  };
};

const toPortalAuditEntry = (entry: AuditLogEntry): AdminAuditLogEntry => ({
  action: entry.action,
  actorLabel: entry.actor_label,
  createdAt: entry.created_at,
  id: entry.id,
});

/**
 * Журнал действий. Основной источник — `back/admin`; если сервис недоступен
 * (или разделы портала пишут события в локальный журнал), берём данные из
 * portal-api, чтобы админка не оставалась пустой.
 */
export const adminAuditLogRequest = async (params: {
  limit?: number;
  offset?: number;
  userId: string;
}): Promise<AdminAuditLogResult> => {
  const fallback = 'Не удалось загрузить журнал действий';
  const query = omitEmptyParams({
    limit: params.limit ?? 50,
    offset: params.offset ?? 0,
  });

  try {
    const response = await fetchWithNetworkFallback(
      `/api/v1/admin/audit-log${query}`,
      { headers: adminHeaders(params.userId) },
      fallback,
    );

    if (!response.ok) {
      throw new AdminApiError(response.status, await parseApiErrorMessage(response, fallback));
    }

    const remote = parseAuditLog(await response.json(), fallback);
    const local = await readPortalAuditLog(params);

    return mergeAuditLogs(remote, local);
  } catch {
    return readPortalAuditLog(params);
  }
};

const readPortalAuditLog = async (params: {
  limit?: number;
  offset?: number;
}): Promise<AdminAuditLogResult> => {
  const query = omitEmptyParams({ limit: params.limit ?? 50, offset: params.offset ?? 0 });
  const result = await portalRequest<{
    items: AuditLogEntry[];
    pagination: { total: number };
  }>({
    fallback: 'Не удалось загрузить журнал действий',
    path: `/audit-log${query}`,
  });

  return {
    items: result.items.map(toPortalAuditEntry),
    total: result.pagination.total,
  };
};

const mergeAuditLogs = (
  remote: AdminAuditLogResult,
  local: AdminAuditLogResult,
): AdminAuditLogResult => {
  const items = [...remote.items, ...local.items].sort((a, b) =>
    b.createdAt.localeCompare(a.createdAt),
  );

  return { items, total: remote.total + local.total };
};

export const writeAuditEntry = async (action: string, actorLabel?: string): Promise<void> => {
  try {
    await portalRequest({
      body: { action, actorLabel },
      fallback: 'Не удалось записать действие в журнал',
      method: 'POST',
      path: '/audit-log',
    });
  } catch {
    // журнал не должен ломать основной сценарий
  }
};
