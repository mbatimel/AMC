import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Promotion } from '@/core/shared/api/promotions';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { listPromotionsRequest } from '@/core/shared/api/promotions';

export const promoMounted = createEvent();

export const fetchActivePromosFx = createEffect(async (): Promise<Promotion[]> => {
  const promotions = await listPromotionsRequest();

  return promotions.filter((promo) => promo.status === 'active');
});

export const $activePromos = createStore<Promotion[]>([]).on(
  fetchActivePromosFx.doneData,
  (_, promotions) => promotions,
);

export const $isPromoPending = fetchActivePromosFx.pending;

export const $promoError = createStore<null | string>(null)
  .on(fetchActivePromosFx, () => null)
  .on(fetchActivePromosFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить акции'),
  );

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> target */

sample({
  clock: promoMounted,
  source: fetchActivePromosFx.pending,
  filter: (pending) => !pending,
  target: fetchActivePromosFx,
});

/* eslint-enable perfectionist/sort-objects */
