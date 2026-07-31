import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Order } from '@/core/shared/api/orders';
import type { CreateSupportRequestPayload, SupportRequest } from '@/core/shared/api/support';

import { $userId } from '@/core/entities/session';
import { listOrdersRequest } from '@/core/shared/api/orders';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { createSupportRequest, listSupportRequests } from '@/core/shared/api/support';
import { toastShown } from '@/core/shared/ui/Toast/model';

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

export const supportOpened = createEvent();
export const supportSubmitted = createEvent<CreateSupportRequestPayload>();
export const supportFormReset = createEvent();

export const fetchSupportOrdersFx = createEffect(async (userId: string) =>
  listOrdersRequest({ limit: 50, offset: 0, userId }),
);

export const fetchMySupportRequestsFx = createEffect(async (userId: string) =>
  listSupportRequests({ userId }),
);

export const submitSupportFx = createEffect(async (payload: CreateSupportRequestPayload) =>
  createSupportRequest(payload),
);

export const $supportOrders = createStore<Order[]>([]).on(
  fetchSupportOrdersFx.doneData,
  (_, result) => result.items,
);

export const $mySupportRequests = createStore<SupportRequest[]>([])
  .on(fetchMySupportRequestsFx.doneData, (_, items) => items)
  .on(submitSupportFx.doneData, (state, created) => [created, ...state]);

export const $isSupportPending = submitSupportFx.pending;

export const $isSupportSent = createStore(false)
  .on(submitSupportFx.done, () => true)
  .reset(supportFormReset);

export const $supportError = createStore<null | string>(null)
  .on(submitSupportFx, () => null)
  .on(submitSupportFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось отправить обращение'),
  )
  .reset(supportFormReset);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: [supportOpened, $userId],
  source: $userId,
  filter: isUserId,
  target: [fetchSupportOrdersFx, fetchMySupportRequestsFx],
});

sample({
  clock: supportSubmitted,
  target: submitSupportFx,
});

sample({
  clock: submitSupportFx.done,
  fn: () => ({ message: 'Обращение отправлено в поддержку', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: submitSupportFx.failData,
  fn: (error) => ({
    message: toDisplayErrorMessage(error, 'Не удалось отправить обращение'),
    tone: 'error' as const,
  }),
  target: toastShown,
});

/* eslint-enable perfectionist/sort-objects */
