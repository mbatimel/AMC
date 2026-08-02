import { createEffect, createEvent, createStore, sample } from 'effector';

import { $cart, addToCartRequested } from '@/core/entities/cart/model';
import { sessionEnded } from '@/core/entities/session';
import { getProductRequest } from '@/core/shared/api/products';
import { getPrimaryProductImageUrl } from '@/core/shared/lib/productImage';

export const productImagesRemembered = createEvent<Record<string, string>>();

export const fetchMissingProductImagesFx = createEffect(
  async ({ known, productIds }: { known: Record<string, string>; productIds: string[] }) => {
    const missing = [...new Set(productIds)].filter((id) => !known[id]);

    if (missing.length === 0) {
      return {};
    }

    const results = await Promise.allSettled(
      missing.map(async (id) => {
        const product = await getProductRequest(id);
        const url = getPrimaryProductImageUrl(product.images);

        return url ? ([id, url] as const) : null;
      }),
    );

    const next: Record<string, string> = {};

    results.forEach((result) => {
      if (result.status !== 'fulfilled' || !result.value) {
        return;
      }

      const [id, url] = result.value;
      next[id] = url;
    });

    return next;
  },
);

export const $productImageById = createStore<Record<string, string>>({})
  .on(productImagesRemembered, (state, patch) => ({ ...state, ...patch }))
  .on(fetchMissingProductImagesFx.doneData, (state, patch) =>
    Object.keys(patch).length === 0 ? state : { ...state, ...patch },
  )
  .reset(sessionEnded);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: addToCartRequested,
  filter: (payload) => typeof payload.imageUrl === 'string' && payload.imageUrl.length > 0,
  fn: (payload) => ({ [payload.productID]: payload.imageUrl as string }),
  target: productImagesRemembered,
});

sample({
  clock: $cart,
  source: $productImageById,
  filter: (known, cart) => cart.items.some((item) => !known[item.product_id]),
  fn: (known, cart) => ({
    known,
    productIds: cart.items.map((item) => item.product_id),
  }),
  target: fetchMissingProductImagesFx,
});

/* eslint-enable perfectionist/sort-objects */
