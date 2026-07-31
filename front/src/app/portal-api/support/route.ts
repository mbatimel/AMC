import type { SupportRequest } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  createId,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type CreateSupportBody = {
  contact?: string;
  clientName?: string;
  orderId?: string;
  severity?: number;
  source?: SupportRequest['source'];
  subject?: string;
  text?: string;
  userId?: string;
};

export const GET = (request: Request): Response => {
  const status = new URL(request.url).searchParams.get('status');
  const userId = new URL(request.url).searchParams.get('userId');
  let items = readPortalState().support_requests;

  if (status) {
    items = items.filter((item) => item.status === status);
  }

  if (userId) {
    items = items.filter((item) => item.user_id === userId);
  }

  return apiOk({ items, pagination: { limit: items.length, offset: 0, total: items.length } });
};

export const POST = async (request: Request): Promise<Response> => {
  const body = await readJsonBody<CreateSupportBody>(request);

  if (!body?.subject?.trim() || !body.text?.trim()) {
    return apiFail(400, 'Заполните тему и описание обращения');
  }

  const created: SupportRequest = {
    answer: '',
    client_name: body.clientName?.trim() ?? '',
    contact: body.contact?.trim() ?? '',
    created_at: new Date().toISOString(),
    id: createId('support'),
    order_id: body.orderId?.trim() ?? '',
    severity: Number(body.severity) || 3,
    source: body.source === 'assistant' ? 'assistant' : 'form',
    status: 'new',
    subject: body.subject.trim(),
    text: body.text.trim(),
    user_id: body.userId?.trim() ?? '',
  };

  updatePortalState((draft) => {
    draft.support_requests.unshift(created);
  });

  appendAuditEntry('Портал', `Получено обращение в поддержку: ${created.subject}`);

  return apiOk(created);
};
