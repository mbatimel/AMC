import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Product, ProductListItem } from '@/core/shared/api/products';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { getProductRequest, listProductsRequest } from '@/core/shared/api/products';

export const productOpened = createEvent<string>();

export const fetchProductFx = createEffect(async (productId: string) =>
  getProductRequest(productId),
);

export const fetchRelatedFx = createEffect(
  async ({ categoryID, productId }: { categoryID: string; productId: string }) => {
    const result = await listProductsRequest({
      categoryID,
      limit: 8,
      offset: 0,
    });

    return result.items.filter((item) => item.id !== productId).slice(0, 4);
  },
);

export const $product = createStore<null | Product>(null)
  .on(fetchProductFx.doneData, (_, product) => product)
  .on(productOpened, (state, productId) => (state?.id === productId ? state : null));

export const $relatedProducts = createStore<ProductListItem[]>([])
  .on(fetchRelatedFx.doneData, (_, items) => items)
  .on(productOpened, () => []);

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

sample({
  clock: fetchProductFx.doneData,
  filter: (product) => typeof product.category_id === 'string' && product.category_id.length > 0,
  fn: (product) => ({
    categoryID: product.category_id as string,
    productId: product.id,
  }),
  target: fetchRelatedFx,
});

/* eslint-enable perfectionist/sort-objects */
