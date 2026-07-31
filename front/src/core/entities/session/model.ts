import { createEffect, createEvent, createStore, sample } from 'effector';

import type { AuthCredentials, AuthUserResponse } from '@/core/shared/api/auth';

import { loginRequest, registerIndividualRequest, registerIpRequest } from '@/core/shared/api/auth';
import { createSignupRequest } from '@/core/shared/api/signupRequests';

import type { RegisterPayload } from './lib/buildRegisterPayload';

import { RegisterType } from './lib/buildRegisterPayload';
import { clearUserIdCookie, readUserIdCookie, writeUserIdCookie } from './lib/cookie';

export const sessionHydrated = createEvent();
export const sessionStarted = createEvent<string>();
export const sessionEnded = createEvent();
export const authErrorCleared = createEvent();

export const loginFx = createEffect<AuthCredentials, AuthUserResponse, Error>(loginRequest);
export const signupFx = createEffect<RegisterPayload, AuthUserResponse, Error>(async (payload) => {
  if (payload.type === RegisterType.Organization) {
    return registerIpRequest(payload.data);
  }

  return registerIndividualRequest(payload.data);
});

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
export const registerModerationFx = createEffect(async (payload: RegisterPayload) => {
  if (payload.type === RegisterType.Organization) {
    await createSignupRequest({
      company: payload.data.shortName ?? payload.data.fullName ?? '',
      contact: payload.data.directorFullName ?? '',
      email: payload.data.email,
      inn: payload.data.inn ?? '',
      phone: payload.data.phone ?? '',
      type: 'organization',
    });

    return;
  }

  await createSignupRequest({
    company: payload.data.fio,
    contact: payload.data.fio,
    email: payload.data.email,
    inn: payload.data.inn ?? '',
    phone: payload.data.phone,
    type: 'individual',
  });
});

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> fn -> target */

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

/* eslint-enable perfectionist/sort-objects */
