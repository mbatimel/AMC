import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Product } from '@/core/shared/api/products';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { getProductRequest } from '@/core/shared/api/products';

export const productOpened = createEvent<string>();

export const fetchProductFx = createEffect(async (productId: string) =>
  getProductRequest(productId),
);

export const $product = createStore<null | Product>(null)
  .on(fetchProductFx.doneData, (_, product) => product)
  .on(productOpened, (state, productId) => (state?.id === productId ? state : null));

export const $productError = createStore<null | string>(null)
  .on(fetchProductFx, () => null)
  .on(fetchProductFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить товар'),
  );

export const $isProductPending = fetchProductFx.pending;

/** id, для которого уже ушёл (или успешно завершён) запрос. */
const $fetchedProductId = createStore<null | string>(null)
  .on(fetchProductFx, (_, productId) => productId)
  .on(fetchProductFx.fail, () => null);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: productOpened,
  source: {
    fetchedId: $fetchedProductId,
    pending: fetchProductFx.pending,
  },
  filter: ({ fetchedId, pending }, productId) =>
    productId.length > 0 && !pending && fetchedId !== productId,
  fn: (_, productId) => productId,
  target: fetchProductFx,
});

/* eslint-enable perfectionist/sort-objects */
