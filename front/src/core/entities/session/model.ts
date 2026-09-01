import { createEffect, createEvent, createStore, sample } from 'effector';

import type { AuthCredentials, AuthUserResponse, RegisterIpPayload } from '@/core/shared/api/auth';

import { loginRequest, registerIpRequest } from '@/core/shared/api/auth';
import { createSignupRequest } from '@/core/shared/api/signupRequests';

import { clearUserIdCookie, readUserIdCookie, writeUserIdCookie } from './lib/cookie';

export const sessionHydrated = createEvent();
export const sessionStarted = createEvent<string>();
export const sessionEnded = createEvent();
export const authErrorCleared = createEvent();

export const loginFx = createEffect<AuthCredentials, AuthUserResponse, Error>(loginRequest);
export const signupFx = createEffect<RegisterIpPayload, AuthUserResponse, Error>(registerIpRequest);

export const $isSessionHydrated = createStore(false).on(sessionHydrated, () => true);

export const $userId = createStore<null | string>(null)
  .on(sessionStarted, (_, userId) => {
    writeUserIdCookie(userId);

    return userId;
  })
  .on(sessionEnded, () => {
    clearUserIdCookie();

    return null;
  })
  /**
   * Cookie есть — восстанавливаем. Пустая cookie не затирает текущую сессию:
   * иначе hydrate в Header/Footer после логина успевал обнулить $userId.
   */
  .on(sessionHydrated, (current) => readUserIdCookie() ?? current);

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

/**
 * Заявка на модерацию (M-05): после успешной регистрации карточка клиента
 * попадает в админку в раздел «Заявки на регистрацию».
 */
export const registerModerationFx = createEffect(async (payload: RegisterIpPayload) => {
  await createSignupRequest({
    company: payload.shortName ?? payload.fullName ?? '',
    contact: payload.directorFullName ?? '',
    email: payload.email,
    inn: payload.inn ?? '',
    phone: payload.phone ?? '',
    type: 'organization',
  });
});

sample({
  clock: loginFx.doneData,
  fn: ({ userID }) => userID,
  target: sessionStarted,
});

sample({
  clock: signupFx.done,
  fn: ({ params }) => params,
  target: registerModerationFx,
});
