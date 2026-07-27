import { createEffect, createEvent, createStore, sample } from 'effector';

import type { AuthCredentials, AuthUserResponse, SignUpPayload } from '@/core/shared/api/auth';

import { loginRequest, signupRequest } from '@/core/shared/api/auth';

import { clearUserIdCookie, readUserIdCookie, writeUserIdCookie } from './lib/cookie';

export const sessionHydrated = createEvent();
export const sessionStarted = createEvent<string>();
export const sessionEnded = createEvent();
export const authErrorCleared = createEvent();

export const loginFx = createEffect<AuthCredentials, AuthUserResponse, Error>(loginRequest);
export const signupFx = createEffect<SignUpPayload, AuthUserResponse, Error>(signupRequest);

const persistSessionFx = createEffect((userId: string) => {
  writeUserIdCookie(userId);
});

const clearSessionFx = createEffect(() => {
  clearUserIdCookie();
});

export const $userId = createStore<null | string>(null)
  .on(sessionStarted, (_, userId) => userId)
  .on(sessionEnded, () => null);

export const $isAuthPending = createStore(false)
  .on(loginFx, () => true)
  .on(signupFx, () => true)
  .on(loginFx.finally, () => false)
  .on(signupFx.finally, () => false);

export const $authError = createStore<null | string>(null)
  .on(loginFx, () => null)
  .on(signupFx, () => null)
  .on(loginFx.failData, (_, error) => error.message)
  .on(signupFx.failData, (_, error) => error.message)
  .on(authErrorCleared, () => null)
  .on(sessionEnded, () => null);

sample({
  clock: sessionHydrated,
  fn: () => readUserIdCookie(),
  target: $userId,
});

sample({
  clock: loginFx.doneData,
  fn: ({ userID }) => userID,
  target: sessionStarted,
});

sample({
  clock: sessionStarted,
  target: persistSessionFx,
});

sample({
  clock: sessionEnded,
  target: clearSessionFx,
});
