import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Order } from '@/core/shared/api/orders';

import { fetchCartFx } from '@/core/entities/cart';
import { ordersListInvalidated } from '@/core/entities/orders';
import { $userId, sessionEnded } from '@/core/entities/session';
import { createOrderRequest } from '@/core/shared/api/orders';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';

import type { CheckoutFormValues } from './lib/types';

import { deliveryTypeShortLabel } from './lib/constants';

export const checkoutSubmitted = createEvent<CheckoutFormValues>();
export const checkoutSuccessClosed = createEvent();

export const createOrderFx = createEffect(
  async ({ form, userId }: { form: CheckoutFormValues; userId: string }) =>
    createOrderRequest({
      comment: form.comment.trim() || undefined,
      contactName: form.contactName.trim(),
      deliveryAddress: form.deliveryAddress.trim(),
      deliveryType: form.deliveryType,
      email: form.email.trim() || undefined,
      phone: form.phone.trim() || undefined,
      userId,
    }),
);

export const $lastOrder = createStore<null | Order>(null)
  .on(createOrderFx.doneData, (_, order) => order)
  .reset([sessionEnded, checkoutSuccessClosed]);

export const $isSuccessOpen = createStore(false)
  .on(createOrderFx.done, () => true)
  .on(checkoutSuccessClosed, () => false)
  .reset(sessionEnded);

export const $isCheckoutPending = createOrderFx.pending;

export const $checkoutError = createStore<null | string>(null)
  .on(checkoutSubmitted, () => null)
  .on(createOrderFx, () => null)
  .on(createOrderFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось оформить заказ'),
  )
  .reset([sessionEnded, checkoutSuccessClosed]);

export const $successOrderView = $lastOrder.map((order) =>
  order
    ? {
        deliveryType: deliveryTypeShortLabel(order.delivery_type),
        number: order.number || order.id.slice(0, 8),
        total: order.total,
      }
    : null,
);

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: checkoutSubmitted,
  source: $userId,
  filter: isUserId,
  fn: (userId, form) => ({ form, userId }),
  target: createOrderFx,
});

sample({
  clock: checkoutSubmitted,
  source: $userId,
  filter: (userId) => userId === null,
  fn: () => 'Войдите, чтобы оформить заказ',
  target: $checkoutError,
});

sample({
  clock: createOrderFx.done,
  source: $userId,
  filter: isUserId,
  target: fetchCartFx,
});

sample({
  clock: createOrderFx.done,
  target: ordersListInvalidated,
});

/* eslint-enable perfectionist/sort-objects */
