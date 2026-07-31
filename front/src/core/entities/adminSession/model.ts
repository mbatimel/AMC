import { createEffect, createEvent, createStore, sample } from 'effector';

import type { AdminSessionResponse } from '@/core/shared/api/admin';

import { adminLoginRequest, adminLogoutRequest } from '@/core/shared/api/admin';

import {
  clearAdminUserIdCookie,
  readAdminUserIdCookie,
  writeAdminUserIdCookie,
} from './lib/cookie';

export const adminSessionHydrated = createEvent();
export const adminSessionStarted = createEvent<string>();
export const adminSessionEnded = createEvent();
export const adminAuthErrorCleared = createEvent();

export const adminLoginFx = createEffect<
  { email: string; password: string },
  AdminSessionResponse,
  Error
>(adminLoginRequest);

export const adminLogoutFx = createEffect(async (userId: null | string) => {
  if (userId) {
    await adminLogoutRequest(userId);
  }
});

export const $isAdminSessionHydrated = createStore(false).on(adminSessionHydrated, () => true);

export const $adminUserId = createStore<null | string>(null)
  .on(adminSessionStarted, (_, userId) => {
    writeAdminUserIdCookie(userId);

    return userId;
  })
  .on(adminSessionEnded, () => {
    clearAdminUserIdCookie();

    return null;
  })
  .on(adminSessionHydrated, (current) => readAdminUserIdCookie() ?? current);

export const $isAdminAuthPending = adminLoginFx.pending;

export const $adminAuthError = createStore<null | string>(null)
  .on(adminLoginFx, () => null)
  .on(adminLoginFx.failData, (_, error) => error.message)
  .on(adminAuthErrorCleared, () => null)
  .on(adminSessionEnded, () => null);

sample({
  clock: adminLoginFx.doneData,
  fn: ({ userID }) => userID,
  target: adminSessionStarted,
});

sample({
  clock: adminSessionEnded,
  source: $adminUserId,
  target: adminLogoutFx,
});
