import type { SignupRequestStatus } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  createId,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type RouteContext = { params: Promise<{ id: string }> };

type PatchSignupBody = {
  rejectReason?: string;
  status: SignupRequestStatus;
};

export const PATCH = async (request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  const payload = await readJsonBody<PatchSignupBody>(request);

  if (!payload?.status) {
    return apiFail(400, 'Не указан новый статус заявки');
  }

  let found = false;

  updatePortalState((draft) => {
    const item = draft.signup_requests.find((request_) => request_.id === id);

    if (!item) {
      return;
    }

    found = true;
    item.status = payload.status;
    item.reject_reason = payload.rejectReason ?? '';

    if (payload.status === 'approved' && !draft.portal_users.some((u) => u.email === item.email)) {
      draft.portal_users.unshift({
        company: item.company,
        contact: item.contact,
        created_at: new Date().toISOString(),
        email: item.email,
        id: createId('portal-user'),
        inn: item.inn,
        is_active: true,
        phone: item.phone,
        role: 'client',
      });
    }
  });

  if (!found) {
    return apiFail(404, 'Заявка не найдена');
  }

  appendAuditEntry(
    'Админ портала',
    payload.status === 'approved' ? `Одобрена заявка ${id}` : `Отклонена заявка ${id}`,
  );

  return apiOk(readPortalState().signup_requests.find((item) => item.id === id));
};
