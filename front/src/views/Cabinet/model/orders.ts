import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type { Order } from '@/core/shared/api/orders';

import { ordersListInvalidated } from '@/core/entities/orders';
import { $userId, sessionEnded } from '@/core/entities/session';
import { listOrdersRequest } from '@/core/shared/api/orders';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';

const ORDERS_PAGE_SIZE = 20;

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

export const cabinetOrdersOpened = createEvent();
export const cabinetOrdersLoadMore = createEvent();
export const cabinetOrderSelected = createEvent<string>();

export const fetchOrdersFx = createEffect(
  async ({ offset, userId }: { offset: number; userId: string }) =>
    listOrdersRequest({ limit: ORDERS_PAGE_SIZE, offset, userId }),
);

export const $orders = createStore<Order[]>([])
  .on(fetchOrdersFx.done, (state, { params, result }) =>
    params.offset === 0 ? result.items : [...state, ...result.items],
  )
  .reset([sessionEnded, ordersListInvalidated]);

export const $ordersTotal = createStore(0)
  .on(fetchOrdersFx.doneData, (_, result) => result.pagination.total)
  .reset([sessionEnded, ordersListInvalidated]);

export const $ordersOffset = createStore(0)
  .on(fetchOrdersFx.done, (_, { params, result }) => params.offset + result.items.length)
  .reset([sessionEnded, ordersListInvalidated]);

export const $selectedOrderId = createStore<null | string>(null)
  .on(cabinetOrderSelected, (_, id) => (id.length > 0 ? id : null))
  .reset(sessionEnded);

export const $selectedOrder = combine($orders, $selectedOrderId, (orders, id) =>
  id ? (orders.find((order) => order.id === id) ?? null) : null,
);

export const $hasMoreOrders = combine(
  $orders,
  $ordersTotal,
  (orders, total) => orders.length < total,
);

export const $isOrdersPending = fetchOrdersFx.pending;

export const $ordersError = createStore<null | string>(null)
  .on(fetchOrdersFx, () => null)
  .on(fetchOrdersFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить заказы'),
  )
  .reset(sessionEnded);

/** Кабинет открывался — ждём userId / рефетч после invalidate. */
const $wantOrders = createStore(false)
  .on(cabinetOrdersOpened, () => true)
  .reset(sessionEnded);

/**
 * Первый page уже ушёл в сеть (или успешно завершён).
 * Сбрасываем при fail offset=0, чтобы можно было повторить.
 */
const $isOrdersFetched = createStore(false)
  .on(fetchOrdersFx, (state, params) => (params.offset === 0 ? true : state))
  .on(fetchOrdersFx.fail, (state, { params }) => (params.offset === 0 ? false : state))
  .reset([sessionEnded, ordersListInvalidated]);

const $initialOrdersPayload = combine(
  $wantOrders,
  $isOrdersFetched,
  fetchOrdersFx.pending,
  $userId,
  (want, fetched, pending, userId) => {
    if (!want || fetched || pending || !isUserId(userId)) {
      return null;
    }

    return { offset: 0, userId };
  },
);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: [cabinetOrdersOpened, $userId, ordersListInvalidated],
  source: $initialOrdersPayload,
  filter: (payload): payload is { offset: number; userId: string } => payload !== null,
  target: fetchOrdersFx,
});

sample({
  clock: cabinetOrdersLoadMore,
  source: {
    hasMore: $hasMoreOrders,
    offset: $ordersOffset,
    pending: $isOrdersPending,
    userId: $userId,
  },
  filter: ({ hasMore, pending, userId }) => isUserId(userId) && !pending && hasMore,
  fn: ({ offset, userId }) => ({ offset, userId: userId as string }),
  target: fetchOrdersFx,
});

/* eslint-enable perfectionist/sort-objects */
