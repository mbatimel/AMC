import { createEffect, createEvent, createStore, sample } from 'effector';

import type { OrderFeedback } from '@/core/shared/api/feedback';
import type { SupportRequest, SupportRequestStatus } from '@/core/shared/api/support';

import { listOrderFeedback } from '@/core/shared/api/feedback';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { listSupportRequests, patchSupportRequest } from '@/core/shared/api/support';
import { toastShown } from '@/core/shared/ui/Toast/model';

export type AdminFeedbackFilters = {
  from: string;
  query: string;
  rating: string;
  to: string;
};

export const DEFAULT_FEEDBACK_FILTERS: AdminFeedbackFilters = {
  from: '',
  query: '',
  rating: '',
  to: '',
};

export const adminFeedbackOpened = createEvent();
export const adminSupportOpened = createEvent();
export const feedbackFiltersChanged = createEvent<Partial<AdminFeedbackFilters>>();
export const supportRequestUpdated = createEvent<{
  answer?: string;
  id: string;
  status: SupportRequestStatus;
}>();

export const fetchAdminFeedbackFx = createEffect(() => listOrderFeedback());
export const fetchAdminSupportFx = createEffect(() => listSupportRequests());

export const updateSupportFx = createEffect(
  async ({ answer, id, status }: { answer?: string; id: string; status: SupportRequestStatus }) =>
    patchSupportRequest(id, { answer, status }),
);

export const $adminFeedback = createStore<OrderFeedback[]>([]).on(
  fetchAdminFeedbackFx.doneData,
  (_, items) => items,
);

export const $adminSupport = createStore<SupportRequest[]>([])
  .on(fetchAdminSupportFx.doneData, (_, items) => items)
  .on(updateSupportFx.doneData, (state, updated) =>
    state.map((item) => (item.id === updated.id ? updated : item)),
  );

export const $feedbackFilters = createStore<AdminFeedbackFilters>(DEFAULT_FEEDBACK_FILTERS).on(
  feedbackFiltersChanged,
  (state, patch) => ({ ...state, ...patch }),
);

export const $isFeedbackPending = fetchAdminFeedbackFx.pending;
export const $isSupportPending = fetchAdminSupportFx.pending;

export const $adminFeedbackError = createStore<null | string>(null)
  .on([fetchAdminFeedbackFx, fetchAdminSupportFx], () => null)
  .on([fetchAdminFeedbackFx.failData, fetchAdminSupportFx.failData], (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить данные'),
  );

sample({
  clock: adminFeedbackOpened,
  target: fetchAdminFeedbackFx,
});

sample({
  clock: adminSupportOpened,
  target: fetchAdminSupportFx,
});

sample({
  clock: supportRequestUpdated,
  target: updateSupportFx,
});

sample({
  clock: updateSupportFx.done,
  fn: () => ({ message: 'Обращение обновлено', tone: 'success' as const }),
  target: toastShown,
});
