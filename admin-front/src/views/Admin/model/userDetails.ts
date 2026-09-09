import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Order } from '@/core/shared/api/orders';
import type { UserBlockPayload } from '@/core/shared/api/userBlock';
import type { RealUser } from '@/core/shared/api/users';

import { listUserOrdersRequest } from '@/core/shared/api/orders';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { getUserRequest, setUserActiveRequest } from '@/core/shared/api/users';

export type ToggleUserDetailPayload = {
  deactivate?: UserBlockPayload;
  id: string;
  isActive: boolean;
};

export const adminUserDetailOpened = createEvent<string>();
export const adminUserDetailStatusToggled = createEvent<ToggleUserDetailPayload>();

export const fetchUserDetailFx = createEffect((userId: string) => getUserRequest(userId));
export const fetchUserOrdersFx = createEffect((userId: string) => listUserOrdersRequest(userId));
export const toggleUserDetailStatusFx = createEffect<ToggleUserDetailPayload, RealUser, Error>(
  async (payload) =>
    setUserActiveRequest({
      deactivate: payload.deactivate,
      isActive: payload.isActive,
      userId: payload.id,
    }),
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
export const $isUserStatusPending = toggleUserDetailStatusFx.pending;

export const $userDetailError = createStore<null | string>(null)
  .on([fetchUserDetailFx, fetchUserOrdersFx, toggleUserDetailStatusFx], () => null)
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
