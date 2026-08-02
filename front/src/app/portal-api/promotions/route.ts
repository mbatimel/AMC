import type { Promotion } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  createId,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

import { type PromotionWriteBody, validatePromotionPayload } from './lib';

export const dynamic = 'force-dynamic';

export const GET = (): Response => {
  const items = readPortalState().promotions ?? [];

  return apiOk({ items });
};

export const POST = async (request: Request): Promise<Response> => {
  const body = await readJsonBody<PromotionWriteBody>(request);
  const validated = validatePromotionPayload(body);

  if ('error' in validated) {
    return apiFail(400, validated.error);
  }

  const created: Promotion = {
    ...validated.value,
    createdAt: new Date().toISOString(),
    endedManually: false,
    id: createId('promo'),
  };

  updatePortalState((draft) => {
    draft.promotions = [created, ...(draft.promotions ?? [])];
  });

  appendAuditEntry('Админ портала', `Создана акция: ${created.name}`);

  return apiOk(created);
};
