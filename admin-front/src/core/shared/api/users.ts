import {
  assertApiSuccess,
  fetchWithNetworkFallback,
  omitEmptyParams,
  parseApiErrorMessage,
} from './parseApiError';

export type ListUsersParams = {
  clientID?: string;
  isActive?: boolean;
  limit?: number;
  offset?: number;
  q?: string;
  role?: string;
  sort?: string;
  status?: string;
};

/** Пользователь реального сервиса `users` (Postgres), а не мок-хранилища portal-api. */
export type RealUser = {
  active_client_id: string;
  client_id: string;
  company_name: string;
  created_at: string;
  email: string;
  first_name: string;
  id: string;
  inn: string;
  is_active: boolean;
  last_name: string;
  middle_name: string;
  phone: string;
  role: string;
  status: string;
  updated_at: string;
};

export class UsersApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'UsersApiError';
    this.status = status;
  }
}

const asString = (value: unknown, fallback = ''): string =>
  typeof value === 'string' ? value : fallback;

const asBoolean = (value: unknown, fallback = false): boolean =>
  typeof value === 'boolean' ? value : fallback;

const parseUser = (value: unknown): null | RealUser => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);

  if (!id) {
    return null;
  }

  return {
    active_client_id: asString(record.active_client_id),
    client_id: asString(record.client_id),
    company_name: asString(record.company_name),
    created_at: asString(record.created_at),
    email: asString(record.email),
    first_name: asString(record.first_name),
    id,
    inn: asString(record.inn),
    is_active: asBoolean(record.is_active, true),
    last_name: asString(record.last_name),
    middle_name: asString(record.middle_name),
    phone: asString(record.phone),
    role: asString(record.role, 'client'),
    status: asString(record.status),
    updated_at: asString(record.updated_at),
  };
};

const request = async (path: string, init?: RequestInit): Promise<unknown> => {
  const fallback = 'Не удалось выполнить запрос к сервису пользователей';
  // Со слэшем: nginx location `/api/v1/users/` резолвит только с завершающим слэшем на
  // коллекции (см. rewrite в next.config.ts и аналогичный комментарий в front/orders.ts).
  const response = await fetchWithNetworkFallback(path, init, fallback);

  if (!response.ok) {
    throw new UsersApiError(response.status, await parseApiErrorMessage(response, fallback));
  }

  const record = assertApiSuccess(await response.json(), fallback);

  return record.data;
};

export const listUsersRequest = async (
  params: ListUsersParams = {},
): Promise<{ items: RealUser[] }> => {
  const query = omitEmptyParams({
    clientID: params.clientID,
    isActive: params.isActive,
    limit: params.limit ?? 200,
    offset: params.offset,
    q: params.q,
    role: params.role,
    sort: params.sort,
    status: params.status,
  });

  const data = await request(`/api/v1/users${query}`);
  const record = typeof data === 'object' && data !== null ? (data as Record<string, unknown>) : {};
  const rawItems = Array.isArray(record.items) ? record.items : [];

  return {
    items: rawItems.flatMap((item) => {
      const parsed = parseUser(item);

      return parsed ? [parsed] : [];
    }),
  };
};

export const getUserRequest = async (userId: string): Promise<RealUser> => {
  const data = await request(`/api/v1/users/${userId}`);
  const user = parseUser(data);

  if (!user) {
    throw new UsersApiError(500, 'Некорректный ответ сервиса пользователей');
  }

  return user;
};

export const setUserActiveRequest = async (
  userId: string,
  isActive: boolean,
): Promise<RealUser> => {
  const data = await request(`/api/v1/users/${userId}/${isActive ? 'activate' : 'deactivate'}`, {
    method: 'POST',
  });
  const user = parseUser(data);

  if (!user) {
    throw new UsersApiError(500, 'Некорректный ответ сервиса пользователей');
  }

  return user;
};
