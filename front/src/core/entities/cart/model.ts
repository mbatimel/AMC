import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type { Cart } from '@/core/shared/api/cart';

import { $userId, sessionEnded } from '@/core/entities/session';
import {
  addCartItemRequest,
  clearCartRequest,
  deleteCartItemRequest,
  getCartRequest,
  updateCartItemRequest,
} from '@/core/shared/api/cart';
import { toastShown } from '@/core/shared/ui/Toast/model';

export type AddToCartPayload = {
  name?: string;
  productID: string;
  qty: number;
};

const emptyCart = (): Cart => ({
  client_id: '',
  discount_total: 0,
  id: '',
  items: [],
  subtotal: 0,
  total: 0,
  user_id: '',
  vat: 0,
});

export const cartHydrated = createEvent();
export const addToCartRequested = createEvent<AddToCartPayload>();
export const cartItemQtyChanged = createEvent<{ cartItemID: string; qty: number }>();
export const cartItemRemoved = createEvent<string>();
export const cartClearRequested = createEvent();

export const fetchCartFx = createEffect(async (userId: string) => getCartRequest({ userId }));

export const addToCartFx = createEffect(
  async ({ productID, qty, userId }: AddToCartPayload & { userId: string }) =>
    addCartItemRequest({ productID, qty, userId }),
);

export const updateCartItemFx = createEffect(
  async ({ cartItemID, qty, userId }: { cartItemID: string; qty: number; userId: string }) =>
    updateCartItemRequest({ cartItemID, qty, userId }),
);

export const deleteCartItemFx = createEffect(
  async ({ cartItemID, userId }: { cartItemID: string; userId: string }) =>
    deleteCartItemRequest({ cartItemID, userId }),
);

export const clearCartFx = createEffect(async (userId: string) => clearCartRequest({ userId }));

export const $cart = createStore<Cart>(emptyCart())
  .on(fetchCartFx.doneData, (_, cart) => cart)
  .on(addToCartFx.doneData, (_, cart) => cart)
  .on(updateCartItemFx.doneData, (_, cart) => cart)
  .on(deleteCartItemFx.doneData, (_, cart) => cart)
  .on(clearCartFx.doneData, (_, cart) => cart)
  .reset(sessionEnded);

/** Уже запрашивали корзину для текущего пользователя. */
export const $isCartFetched = createStore(false)
  .on(fetchCartFx, () => true)
  .reset(sessionEnded);

export const $cartCount = $cart.map((cart) => cart.items.length);

export const $isCartPending = combine(
  fetchCartFx.pending,
  addToCartFx.pending,
  updateCartItemFx.pending,
  deleteCartItemFx.pending,
  clearCartFx.pending,
  (...flags) => flags.some(Boolean),
);

export const $cartError = createStore<null | string>(null)
  .on(fetchCartFx, () => null)
  .on(addToCartFx, () => null)
  .on(fetchCartFx.failData, (_, error) => error.message)
  .on(addToCartFx.failData, (_, error) => error.message)
  .on(updateCartItemFx.failData, (_, error) => error.message)
  .on(deleteCartItemFx.failData, (_, error) => error.message)
  .on(clearCartFx.failData, (_, error) => error.message)
  .reset(sessionEnded);

const $canFetchCart = combine(
  fetchCartFx.pending,
  $isCartFetched,
  (pending, fetched) => !pending && !fetched,
);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

/** Один GET при появлении userId (hydrate сессии / логин). */
sample({
  clock: $userId,
  source: combine($canFetchCart, $userId, (canFetch, userId) =>
    canFetch && typeof userId === 'string' && userId.length > 0 ? userId : null,
  ),
  filter: (userId): userId is string => typeof userId === 'string',
  target: fetchCartFx,
});

/** Ручной hydrate — только если ещё не грузили. */
sample({
  clock: cartHydrated,
  source: combine($canFetchCart, $userId, (canFetch, userId) =>
    canFetch && typeof userId === 'string' && userId.length > 0 ? userId : null,
  ),
  filter: (userId): userId is string => typeof userId === 'string',
  target: fetchCartFx,
});

sample({
  clock: addToCartRequested,
  source: $userId,
  filter: (userId: null | string): userId is string =>
    typeof userId === 'string' && userId.length > 0,
  fn: (userId, payload) => ({ ...payload, userId }),
  target: addToCartFx,
});

sample({
  clock: cartItemQtyChanged,
  source: $userId,
  filter: (userId: null | string): userId is string =>
    typeof userId === 'string' && userId.length > 0,
  fn: (userId, payload) => ({ ...payload, userId }),
  target: updateCartItemFx,
});

sample({
  clock: cartItemRemoved,
  source: $userId,
  filter: (userId: null | string): userId is string =>
    typeof userId === 'string' && userId.length > 0,
  fn: (userId, cartItemID) => ({ cartItemID, userId }),
  target: deleteCartItemFx,
});

sample({
  clock: cartClearRequested,
  source: $userId,
  filter: (userId: null | string): userId is string =>
    typeof userId === 'string' && userId.length > 0,
  target: clearCartFx,
});

sample({
  clock: addToCartRequested,
  source: $userId,
  filter: (userId) => userId === null,
  fn: () => ({
    message: 'Войдите, чтобы добавить товар в корзину',
    tone: 'error' as const,
  }),
  target: toastShown,
});

sample({
  clock: addToCartFx.done,
  fn: ({ params }) => ({
    message: params.name
      ? `Добавлено в корзину: ${params.name} × ${params.qty}`
      : `Добавлено в корзину × ${params.qty}`,
    tone: 'success' as const,
  }),
  target: toastShown,
});

sample({
  clock: addToCartFx.failData,
  fn: (error) => ({
    message: error.message,
    tone: 'error' as const,
  }),
  target: toastShown,
});
