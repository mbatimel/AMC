import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Brand } from '@/core/shared/api/products';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { listBrandsRequest } from '@/core/shared/api/products';

export const brandsMounted = createEvent();

export const fetchBrandsFx = createEffect(() => listBrandsRequest());

export const $brands = createStore<Brand[]>([]).on(fetchBrandsFx.doneData, (_, brands) => brands);

export const $isBrandsPending = fetchBrandsFx.pending;

export const $brandsError = createStore<null | string>(null)
  .on(fetchBrandsFx, () => null)
  .on(fetchBrandsFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить бренды'),
  );

const $isBrandsLoaded = createStore(false).on(fetchBrandsFx.done, () => true);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> target */

sample({
  clock: brandsMounted,
  source: { loaded: $isBrandsLoaded, pending: fetchBrandsFx.pending },
  filter: ({ loaded, pending }) => !loaded && !pending,
  target: fetchBrandsFx,
});

/* eslint-enable perfectionist/sort-objects */
