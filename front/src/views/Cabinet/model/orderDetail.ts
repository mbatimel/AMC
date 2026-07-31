import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Order } from '@/core/shared/api/orders';

import { $userId, sessionEnded } from '@/core/entities/session';
import { getOrderRequest } from '@/core/shared/api/orders';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

export const cabinetOrderOpened = createEvent<string>();
export const cabinetOrderClosed = createEvent();

export const fetchOrderFx = createEffect(
  async ({ orderID, userId }: { orderID: string; userId: string }) =>
    getOrderRequest({ orderID, userId }),
);

export const $orderDetailId = createStore<null | string>(null)
  .on(cabinetOrderOpened, (_, orderID) => orderID)
  .reset([sessionEnded, cabinetOrderClosed]);

export const $orderDetail = createStore<null | Order>(null)
  .on(fetchOrderFx.doneData, (_, order) => order)
  .reset([sessionEnded, cabinetOrderClosed, cabinetOrderOpened]);

export const $isOrderDetailPending = fetchOrderFx.pending;

export const $orderDetailError = createStore<null | string>(null)
  .on(cabinetOrderOpened, () => null)
  .on(fetchOrderFx, () => null)
  .on(fetchOrderFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить заказ'),
  )
  .reset([sessionEnded, cabinetOrderClosed]);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: cabinetOrderOpened,
  source: $userId,
  filter: isUserId,
  fn: (userId, orderID) => ({ orderID, userId }),
  target: fetchOrderFx,
});

sample({
  clock: $userId.updates,
  source: {
    orderID: $orderDetailId,
    pending: $isOrderDetailPending,
  },
  filter: ({ orderID, pending }, userId) =>
    isUserId(userId) && typeof orderID === 'string' && orderID.length > 0 && !pending,
  fn: ({ orderID }, userId) => ({ orderID: orderID as string, userId: userId as string }),
  target: fetchOrderFx,
});

/* eslint-enable perfectionist/sort-objects */
