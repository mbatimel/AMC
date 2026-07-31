import type { SupportRequestStatus } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import { appendAuditEntry, readPortalState, updatePortalState } from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type RouteContext = { params: Promise<{ id: string }> };

type PatchSupportBody = {
  answer?: string;
  status?: SupportRequestStatus;
};

export const PATCH = async (request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  const payload = await readJsonBody<PatchSupportBody>(request);

  if (!payload) {
    return apiFail(400, 'Некорректное тело запроса');
  }

  let found = false;

  updatePortalState((draft) => {
    const item = draft.support_requests.find((request_) => request_.id === id);

    if (!item) {
      return;
    }

    found = true;
    item.status = payload.status ?? item.status;
    item.answer = payload.answer ?? item.answer;
  });

  if (!found) {
    return apiFail(404, 'Обращение не найдено');
  }

  appendAuditEntry('Админ портала', `Обновлено обращение поддержки ${id}`);

  return apiOk(readPortalState().support_requests.find((item) => item.id === id));
};
