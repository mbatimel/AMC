import type {
  Promotion,
  PromotionDiscMode,
  PromotionSelection,
  PromotionType,
} from '@/core/shared/server/portal/types';

import { portalRequest } from './portalClient';

export type {
  Promotion,
  PromotionDiscMode,
  PromotionSelection,
  PromotionType,
} from '@/core/shared/server/portal/types';

export type PromotionWritePayload = {
  condition?: string;
  desc?: string;
  discMode: PromotionDiscMode;
  discValue: number;
  endAt: string;
  minQty?: number;
  name: string;
  sel: PromotionSelection;
  startAt: string;
  type: PromotionType;
};

export const listPromotions = async (): Promise<Promotion[]> => {
  const result = await portalRequest<{ items: Promotion[] }>({
    fallback: 'Не удалось загрузить акции',
    path: '/promotions',
  });

  return result.items;
};

export const getPromotion = (id: string): Promise<Promotion> =>
  portalRequest({
    fallback: 'Не удалось загрузить акцию',
    path: `/promotions/${id}`,
  });

export const createPromotion = (payload: PromotionWritePayload): Promise<Promotion> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось создать акцию',
    method: 'POST',
    path: '/promotions',
  });

export const updatePromotion = (id: string, payload: PromotionWritePayload): Promise<Promotion> =>
  portalRequest({
    body: { ...payload, endedManually: false },
    fallback: 'Не удалось сохранить акцию',
    method: 'PATCH',
    path: `/promotions/${id}`,
  });

export const endPromotion = (id: string): Promise<Promotion> =>
  portalRequest({
    body: { endedManually: true },
    fallback: 'Не удалось завершить акцию',
    method: 'PATCH',
    path: `/promotions/${id}`,
  });

export const deletePromotion = (id: string): Promise<{ id: string }> =>
  portalRequest({
    fallback: 'Не удалось удалить акцию',
    method: 'DELETE',
    path: `/promotions/${id}`,
  });
