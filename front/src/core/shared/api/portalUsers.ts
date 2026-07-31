import type { PortalUser } from '@/core/shared/server/portal/types';

import { portalRequest } from './portalClient';

export type { PortalUser } from '@/core/shared/server/portal/types';

export const listPortalUsers = async (): Promise<PortalUser[]> => {
  const result = await portalRequest<{ items: PortalUser[] }>({
    fallback: 'Не удалось загрузить пользователей портала',
    path: '/portal-users',
  });

  return result.items;
};

export const invitePortalUser = (payload: {
  company?: string;
  contact?: string;
  email: string;
  role: PortalUser['role'];
}): Promise<PortalUser> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось отправить приглашение',
    method: 'POST',
    path: '/portal-users',
  });

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
