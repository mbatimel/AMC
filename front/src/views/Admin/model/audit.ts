import { createEffect, createEvent, createStore, sample } from 'effector';

import type { AdminAuditLogEntry } from '@/core/shared/api/admin';

import { $adminUserId } from '@/core/entities/adminSession';
import { adminAuditLogRequest } from '@/core/shared/api/admin';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

export const adminAuditLogOpened = createEvent();

export const fetchAuditLogFx = createEffect(async (userId: string) =>
  adminAuditLogRequest({ limit: 100, offset: 0, userId }),
);

export const $auditLog = createStore<AdminAuditLogEntry[]>([]).on(
  fetchAuditLogFx.doneData,
  (_, result) => result.items,
);

export const $auditLogTotal = createStore(0).on(
  fetchAuditLogFx.doneData,
  (_, result) => result.total,
);

export const $isAuditLogPending = fetchAuditLogFx.pending;

export const $auditLogError = createStore<null | string>(null)
  .on(fetchAuditLogFx, () => null)
  .on(fetchAuditLogFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить журнал действий'),
  );

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> target */

sample({
  clock: [adminAuditLogOpened, $adminUserId],
  source: $adminUserId,
  filter: isUserId,
  target: fetchAuditLogFx,
});

/* eslint-enable perfectionist/sort-objects */
