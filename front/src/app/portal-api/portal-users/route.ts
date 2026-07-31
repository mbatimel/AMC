import type { PortalUser } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  createId,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type CreatePortalUserBody = {
  company?: string;
  contact?: string;
  email?: string;
  role?: PortalUser['role'];
};

export const GET = (): Response => apiOk({ items: readPortalState().portal_users });

export const POST = async (request: Request): Promise<Response> => {
  const body = await readJsonBody<CreatePortalUserBody>(request);

  if (!body?.email?.trim()) {
    return apiFail(400, 'Укажите e-mail');
  }

  const created: PortalUser = {
    company: body.company?.trim() || 'Администратор портала',
    contact: body.contact?.trim() ?? '',
    created_at: new Date().toISOString(),
    email: body.email.trim(),
    id: createId('portal-user'),
    inn: '',
    is_active: true,
    phone: '',
    role: body.role === 'admin' ? 'admin' : 'client',
  };

  updatePortalState((draft) => {
    draft.portal_users.unshift(created);
  });

  appendAuditEntry('Админ портала', `Приглашён пользователь портала: ${created.email}`);

  return apiOk(created);
};
