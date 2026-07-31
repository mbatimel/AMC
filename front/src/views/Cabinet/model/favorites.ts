import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type { Product } from '@/core/shared/api/products';

import { $favoriteIds } from '@/core/entities/favorites';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { getProductRequest } from '@/core/shared/api/products';

export const cabinetFavoritesOpened = createEvent();

export const fetchFavoriteProductsFx = createEffect(async (ids: string[]): Promise<Product[]> => {
  const results = await Promise.allSettled(ids.map((id) => getProductRequest(id)));

  return results.flatMap((result) => (result.status === 'fulfilled' ? [result.value] : []));
});

export const $favoriteProducts = createStore<Product[]>([]).on(
  fetchFavoriteProductsFx.doneData,
  (_, products) => products,
);

export const $isFavoritesPending = fetchFavoriteProductsFx.pending;

export const $favoritesError = createStore<null | string>(null)
  .on(fetchFavoriteProductsFx, () => null)
  .on(fetchFavoriteProductsFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить избранное'),
  );

/** Показываем только те товары, которые всё ещё в избранном. */
export const $visibleFavorites = combine($favoriteProducts, $favoriteIds, (products, ids) =>
  products.filter((product) => ids.includes(product.id)),
);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> target */

sample({
  clock: [cabinetFavoritesOpened, $favoriteIds],
  source: $favoriteIds,
  filter: (ids) => ids.length > 0,
  target: fetchFavoriteProductsFx,
});

/* eslint-enable perfectionist/sort-objects */
