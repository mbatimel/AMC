import type { PortalUser } from '@/core/shared/server/portal/types';

import { assertApiSuccess, fetchWithNetworkFallback, parseApiErrorMessage } from './parseApiError';
import { portalRequest } from './portalClient';

export type { PortalUser } from '@/core/shared/server/portal/types';

export type InvitePortalAdminPayload = {
  email: string;
  name: string;
};

export type InvitePortalAdminResult = {
  email: string;
  emailSent: boolean;
  password: string;
  userId: string;
};

export class PortalUsersApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'PortalUsersApiError';
    this.status = status;
  }
}

export const listPortalUsers = async (): Promise<PortalUser[]> => {
  const result = await portalRequest<{ items: PortalUser[] }>({
    fallback: 'Не удалось загрузить пользователей портала',
    path: '/portal-users',
  });

  return result.items;
};

const localizeInviteError = (status: number, message: string): string => {
  const normalized = message.trim().toLowerCase();

  if (status === 409 || normalized.includes('conflict') || normalized.includes('already')) {
    return 'Пользователь с таким e-mail уже существует';
  }

  if (status === 400 || normalized.includes('bad request') || normalized.includes('validation')) {
    return 'Проверьте e-mail и имя';
  }

  return message.trim().length > 0 ? message : 'Не удалось отправить приглашение';
};

/**
 * Приглашение администратора.
 * POST /api/v1/admin/portal-users/invite — { email, name }
 * В ответе всегда есть password (fallback, если письмо не ушло).
 */
export const invitePortalUser = async (
  adminUserId: string,
  payload: InvitePortalAdminPayload,
): Promise<InvitePortalAdminResult> => {
  const fallback = 'Не удалось отправить приглашение';
  const response = await fetchWithNetworkFallback(
    '/api/v1/admin/portal-users/invite',
    {
      body: JSON.stringify({
        email: payload.email,
        name: payload.name,
      }),
      headers: {
        'Content-Type': 'application/json',
        'X-User-Id': adminUserId,
      },
      method: 'POST',
    },
    fallback,
  );

  if (!response.ok) {
    const raw = await parseApiErrorMessage(response, fallback);

    throw new PortalUsersApiError(response.status, localizeInviteError(response.status, raw));
  }

  const record = assertApiSuccess(await response.json(), fallback);
  const data = record.data;

  if (typeof data !== 'object' || data === null) {
    throw new PortalUsersApiError(500, fallback);
  }

  const source = data as Record<string, unknown>;
  const userId = typeof source.userId === 'string' ? source.userId : '';
  const email = typeof source.email === 'string' ? source.email : payload.email;
  const password = typeof source.password === 'string' ? source.password : '';

  if (!userId || !password) {
    throw new PortalUsersApiError(500, 'Некорректный ответ сервиса приглашений');
  }

  return {
    email,
    emailSent: source.emailSent === true,
    password,
    userId,
  };
};

export const patchPortalUser = (
  id: string,
  payload: { isActive?: boolean; passwordReset?: boolean },
): Promise<PortalUser> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось обновить пользователя',
    method: 'PATCH',
    path: `/portal-users/${id}`,
  });
