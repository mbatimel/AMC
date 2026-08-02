import { createEffect, createEvent, createStore, sample } from 'effector';

import type { AdminSessionResponse } from '@/core/shared/api/admin';

import {
  adminLoginRequest,
  adminLogoutRequest,
  adminSessionRequest,
} from '@/core/shared/api/admin';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';

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

export const adminHydrateFx = createEffect(async (): Promise<null | string> => {
  const userId = readAdminUserIdCookie();

  if (!userId) {
    return null;
  }

  try {
    const session = await adminSessionRequest(userId);

    return session.userID;
  } catch (error) {
    clearAdminUserIdCookie();
    throw error;
  }
});

export const adminLogoutFx = createEffect(async (userId: null | string) => {
  if (userId) {
    await adminLogoutRequest(userId);
  }
});

export const $isAdminSessionHydrated = createStore(false).on(adminHydrateFx.finally, () => true);

export const $adminUserId = createStore<null | string>(null)
  .on(adminSessionStarted, (_, userId) => {
    writeAdminUserIdCookie(userId);

    return userId;
  })
  .on(adminSessionEnded, () => {
    clearAdminUserIdCookie();

    return null;
  })
  /**
   * `null` из hydrate (нет cookie) не затирает свежий login:
   * иначе гонка hydrate → login → hydrate.done(null) обнуляла сессию.
   */
  .on(adminHydrateFx.doneData, (current, userId) => userId ?? current);

export const $isAdminAuthPending = adminLoginFx.pending;

export const $adminAuthError = createStore<null | string>(null)
  .on(adminLoginFx, () => null)
  .on(adminLoginFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось выполнить вход'),
  )
  .on(adminAuthErrorCleared, () => null)
  .on(adminSessionEnded, () => null);

sample({
  clock: adminSessionHydrated,
  target: adminHydrateFx,
});

sample({
  clock: adminLoginFx.doneData,
  fn: ({ userID }) => userID,
  target: adminSessionStarted,
});

sample({
  clock: adminHydrateFx.fail,
  target: adminSessionEnded,
});

sample({
  clock: adminLogoutFx.finally,
  target: adminSessionEnded,
});
