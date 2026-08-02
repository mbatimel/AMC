import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Order } from '@/core/shared/api/orders';
import type { RealUser } from '@/core/shared/api/users';

import { listUserOrdersRequest } from '@/core/shared/api/orders';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { getUserRequest, setUserActiveRequest } from '@/core/shared/api/users';

export const adminUserDetailOpened = createEvent<string>();
export const adminUserDetailStatusToggled = createEvent<{ id: string; isActive: boolean }>();

export const fetchUserDetailFx = createEffect((userId: string) => getUserRequest(userId));
export const fetchUserOrdersFx = createEffect((userId: string) => listUserOrdersRequest(userId));
export const toggleUserDetailStatusFx = createEffect(
  async ({ id, isActive }: { id: string; isActive: boolean }) => setUserActiveRequest(id, isActive),
);

export const $userDetail = createStore<null | RealUser>(null)
  .on(fetchUserDetailFx.doneData, (_, user) => user)
  .on(toggleUserDetailStatusFx.doneData, (_, user) => user)
  .reset(adminUserDetailOpened);

export const $userOrders = createStore<Order[]>([])
  .on(fetchUserOrdersFx.doneData, (_, result) => result.items)
  .reset(adminUserDetailOpened);

export const $isUserDetailPending = fetchUserDetailFx.pending;
export const $isUserOrdersPending = fetchUserOrdersFx.pending;

export const $userDetailError = createStore<null | string>(null)
  .on([fetchUserDetailFx, fetchUserOrdersFx], () => null)
  .on(
    [fetchUserDetailFx.failData, fetchUserOrdersFx.failData, toggleUserDetailStatusFx.failData],
    (_, error) => toDisplayErrorMessage(error, 'Не удалось загрузить данные пользователя'),
  );

sample({
  clock: adminUserDetailOpened,
  target: [fetchUserDetailFx, fetchUserOrdersFx],
});

sample({
  clock: adminUserDetailStatusToggled,
  target: toggleUserDetailStatusFx,
});
