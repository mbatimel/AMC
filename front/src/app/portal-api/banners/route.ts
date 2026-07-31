import type { BannersSettings } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import { appendAuditEntry, readPortalState, updatePortalState } from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

export const GET = (): Response => apiOk(readPortalState().banners);

export const PUT = async (request: Request): Promise<Response> => {
  const body = await readJsonBody<BannersSettings>(request);

  if (!body || !Array.isArray(body.items)) {
    return apiFail(400, 'Некорректное тело запроса');
  }

  const state = updatePortalState((draft) => {
    draft.banners = {
      delay_sec: Number(body.delay_sec) > 0 ? Number(body.delay_sec) : 6,
      items: body.items.map((item, index) => ({ ...item, sort_order: index + 1 })),
    };
  });

  appendAuditEntry('Админ портала', `Обновлены баннеры главной (${body.items.length} шт.)`);

  return apiOk(state.banners);
};
