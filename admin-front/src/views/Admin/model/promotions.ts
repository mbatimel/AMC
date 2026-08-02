import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Category, ProductListItem } from '@/core/shared/api/products';
import type { Promotion, PromotionWritePayload } from '@/core/shared/api/promotions';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { fetchAllProductsRequest, listCategoriesRequest } from '@/core/shared/api/products';
import {
  createPromotion,
  deletePromotion,
  endPromotion,
  listPromotions,
  updatePromotion,
} from '@/core/shared/api/promotions';
import { toastShown } from '@/core/shared/ui/Toast/model';

export const adminPromotionsOpened = createEvent();
export const promotionSaveRequested = createEvent<{
  id: null | string;
  payload: PromotionWritePayload;
}>();
export const promotionEndRequested = createEvent<string>();
export const promotionDeleteRequested = createEvent<string>();

export const fetchPromotionsFx = createEffect(() => listPromotions());

export const fetchPromoCatalogFx = createEffect(async () => {
  const [categories, products] = await Promise.all([
    listCategoriesRequest(),
    fetchAllProductsRequest(),
  ]);

  return { categories, products };
});

export const savePromotionFx = createEffect(
  async ({ id, payload }: { id: null | string; payload: PromotionWritePayload }) => {
    if (id) {
      return updatePromotion(id, payload);
    }

    return createPromotion(payload);
  },
);

export const endPromotionFx = createEffect((id: string) => endPromotion(id));
export const deletePromotionFx = createEffect((id: string) => deletePromotion(id));

export const $promotions = createStore<Promotion[]>([])
  .on(fetchPromotionsFx.doneData, (_, items) => items)
  .on(savePromotionFx.doneData, (state, item) => {
    const index = state.findIndex((promo) => promo.id === item.id);

    if (index === -1) {
      return [item, ...state];
    }

    return state.map((promo) => (promo.id === item.id ? item : promo));
  })
  .on(endPromotionFx.doneData, (state, item) =>
    state.map((promo) => (promo.id === item.id ? item : promo)),
  )
  .on(deletePromotionFx.doneData, (state, result) =>
    state.filter((promo) => promo.id !== result.id),
  );

export const $promoCategories = createStore<Category[]>([]).on(
  fetchPromoCatalogFx.doneData,
  (_, data) => data.categories,
);

export const $promoProducts = createStore<ProductListItem[]>([]).on(
  fetchPromoCatalogFx.doneData,
  (_, data) => data.products,
);

export const $isPromotionsPending = fetchPromotionsFx.pending;
export const $isPromoCatalogPending = fetchPromoCatalogFx.pending;
export const $isPromotionSaving = savePromotionFx.pending;

export const $promotionsError = createStore<null | string>(null)
  .on(
    [fetchPromotionsFx, fetchPromoCatalogFx, savePromotionFx, endPromotionFx, deletePromotionFx],
    () => null,
  )
  .on(
    [
      fetchPromotionsFx.failData,
      fetchPromoCatalogFx.failData,
      savePromotionFx.failData,
      endPromotionFx.failData,
      deletePromotionFx.failData,
    ],
    (_, error) => toDisplayErrorMessage(error, 'Не удалось выполнить операцию с акциями'),
  );

sample({
  clock: adminPromotionsOpened,
  target: [fetchPromotionsFx, fetchPromoCatalogFx],
});

sample({
  clock: promotionSaveRequested,
  target: savePromotionFx,
});

sample({
  clock: promotionEndRequested,
  target: endPromotionFx,
});

sample({
  clock: promotionDeleteRequested,
  target: deletePromotionFx,
});

sample({
  clock: savePromotionFx.done,
  fn: ({ params }) => ({
    message: params.id ? 'Акция сохранена' : 'Акция создана',
    tone: 'success' as const,
  }),
  target: toastShown,
});

sample({
  clock: endPromotionFx.done,
  fn: () => ({ message: 'Акция завершена', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: deletePromotionFx.done,
  fn: () => ({ message: 'Акция удалена', tone: 'success' as const }),
  target: toastShown,
});
