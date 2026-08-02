import type { Promotion } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

import { validatePromotionPayload } from '../lib';

export const dynamic = 'force-dynamic';

type PatchPromotionBody = Partial<{
  condition: string;
  desc: string;
  discMode: Promotion['discMode'];
  discValue: number;
  endAt: string;
  endedManually: boolean;
  minQty: number;
  name: string;
  sel: Promotion['sel'];
  startAt: string;
  type: Promotion['type'];
}>;

type RouteContext = { params: Promise<{ id: string }> };

export const GET = async (_request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  const item = (readPortalState().promotions ?? []).find((promo) => promo.id === id);

  if (!item) {
    return apiFail(404, 'Акция не найдена');
  }

  return apiOk(item);
};

export const PATCH = async (request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  const payload = await readJsonBody<PatchPromotionBody>(request);

  if (!payload) {
    return apiFail(400, 'Некорректное тело запроса');
  }

  const current = (readPortalState().promotions ?? []).find((promo) => promo.id === id);

  if (!current) {
    return apiFail(404, 'Акция не найдена');
  }

  if (payload.endedManually === true && Object.keys(payload).length === 1) {
    updatePortalState((draft) => {
      const item = (draft.promotions ?? []).find((promo) => promo.id === id);

      if (item) {
        item.endedManually = true;
      }
    });

    appendAuditEntry('Админ портала', `Досрочно завершена акция: ${current.name}`);

    return apiOk((readPortalState().promotions ?? []).find((promo) => promo.id === id));
  }

  const validated = validatePromotionPayload({
    condition: payload.condition ?? current.condition,
    desc: payload.desc ?? current.desc,
    discMode: payload.discMode ?? current.discMode,
    discValue: payload.discValue ?? current.discValue,
    endAt: payload.endAt ?? current.endAt,
    minQty: payload.minQty ?? current.minQty,
    name: payload.name ?? current.name,
    sel: payload.sel ?? current.sel,
    startAt: payload.startAt ?? current.startAt,
    type: payload.type ?? current.type,
  });

  if ('error' in validated) {
    return apiFail(400, validated.error);
  }

  updatePortalState((draft) => {
    const item = (draft.promotions ?? []).find((promo) => promo.id === id);

    if (!item) {
      return;
    }

    Object.assign(item, validated.value, {
      endedManually: payload.endedManually ?? false,
    });
  });

  appendAuditEntry('Админ портала', `Изменена акция: ${validated.value.name}`);

  return apiOk((readPortalState().promotions ?? []).find((promo) => promo.id === id));
};

export const DELETE = async (_request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  let removedName = '';

  updatePortalState((draft) => {
    const list = draft.promotions ?? [];
    const item = list.find((promo) => promo.id === id);

    if (!item) {
      return;
    }

    removedName = item.name;
    draft.promotions = list.filter((promo) => promo.id !== id);
  });

  if (!removedName) {
    return apiFail(404, 'Акция не найдена');
  }

  appendAuditEntry('Админ портала', `Удалена акция: ${removedName}`);

  return apiOk({ id });
};
