import { createEffect, createEvent, createStore, sample } from 'effector';

import type { CreateFeedbackPayload, OrderFeedback } from '@/core/shared/api/feedback';

import { $userId } from '@/core/entities/session';
import { createOrderFeedback, listOrderFeedback } from '@/core/shared/api/feedback';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { toastShown } from '@/core/shared/ui/Toast/model';

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

export const orderFeedbackRequested = createEvent();
export const orderFeedbackSubmitted = createEvent<CreateFeedbackPayload>();

export const fetchOrderFeedbackFx = createEffect(async (userId: string) =>
  listOrderFeedback({ userId }),
);

export const submitOrderFeedbackFx = createEffect(async (payload: CreateFeedbackPayload) =>
  createOrderFeedback(payload),
);

export const $orderFeedback = createStore<OrderFeedback[]>([])
  .on(fetchOrderFeedbackFx.doneData, (_, items) => items)
  .on(submitOrderFeedbackFx.doneData, (state, created) => [
    created,
    ...state.filter((item) => item.order_id !== created.order_id),
  ]);

export const $isFeedbackPending = submitOrderFeedbackFx.pending;

export const $feedbackError = createStore<null | string>(null)
  .on(submitOrderFeedbackFx, () => null)
  .on(submitOrderFeedbackFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось отправить отзыв'),
  );

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: [orderFeedbackRequested, $userId],
  source: $userId,
  filter: isUserId,
  target: fetchOrderFeedbackFx,
});

sample({
  clock: orderFeedbackSubmitted,
  target: submitOrderFeedbackFx,
});

sample({
  clock: submitOrderFeedbackFx.done,
  fn: () => ({ message: 'Спасибо! Отзыв сохранён', tone: 'success' as const }),
  target: toastShown,
});

/* eslint-enable perfectionist/sort-objects */
