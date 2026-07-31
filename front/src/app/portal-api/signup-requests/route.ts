import type { SignupRequest } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  createId,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type CreateSignupBody = {
  company?: string;
  contact?: string;
  email?: string;
  inn?: string;
  phone?: string;
  type?: SignupRequest['type'];
};

export const GET = (request: Request): Response => {
  const status = new URL(request.url).searchParams.get('status');
  const items = readPortalState().signup_requests;

  return apiOk({ items: status ? items.filter((item) => item.status === status) : items });
};

export const POST = async (request: Request): Promise<Response> => {
  const body = await readJsonBody<CreateSignupBody>(request);

  if (!body?.email?.trim()) {
    return apiFail(400, 'Не указан e-mail заявки');
  }

  const created: SignupRequest = {
    company: body.company?.trim() ?? '',
    contact: body.contact?.trim() ?? '',
    created_at: new Date().toISOString(),
    email: body.email.trim(),
    id: createId('signup'),
    inn: body.inn?.trim() ?? '',
    phone: body.phone?.trim() ?? '',
    reject_reason: '',
    status: 'pending',
    type: body.type === 'individual' ? 'individual' : 'organization',
  };

  updatePortalState((draft) => {
    draft.signup_requests.unshift(created);
  });

  appendAuditEntry('Портал', `Новая заявка на регистрацию: ${created.company || created.email}`);

  return apiOk(created);
};
