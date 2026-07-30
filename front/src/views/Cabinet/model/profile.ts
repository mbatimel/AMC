import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type {
  ClientConditions,
  Profile,
  ProfileClient,
  ProfileClientListItem,
  UpdateProfilePayload,
} from '@/core/shared/api/profile';

import { $userId, sessionEnded } from '@/core/entities/session';
import { changePasswordRequest } from '@/core/shared/api/auth';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import {
  activateClientRequest,
  getClientConditionsRequest,
  getClientDetailsRequest,
  getProfileRequest,
  listUserClientsRequest,
  updateProfileRequest,
} from '@/core/shared/api/profile';
import { toastShown } from '@/core/shared/ui/Toast/model';

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

export const cabinetProfileOpened = createEvent();
export const profileSaveRequested = createEvent<UpdateProfilePayload>();
export const clientActivateRequested = createEvent<string>();
export const passwordChangeRequested = createEvent<{
  newPassword: string;
  oldPassword: string;
}>();

export const fetchProfileFx = createEffect(async (userId: string) => getProfileRequest(userId));

export const fetchClientsFx = createEffect(async (userId: string) =>
  listUserClientsRequest(userId),
);

export const fetchClientDetailsFx = createEffect(
  async ({ clientID, userId }: { clientID: string; userId: string }) =>
    getClientDetailsRequest(userId, clientID),
);

export const fetchConditionsFx = createEffect(
  async ({ clientID, userId }: { clientID: string; userId: string }) =>
    getClientConditionsRequest(userId, clientID),
);

export const updateProfileFx = createEffect(
  async ({ payload, userId }: { payload: UpdateProfilePayload; userId: string }) =>
    updateProfileRequest(userId, payload),
);

export const activateClientFx = createEffect(
  async ({ clientID, userId }: { clientID: string; userId: string }) =>
    activateClientRequest(userId, clientID),
);

export const changePasswordFx = createEffect(
  async ({
    newPassword,
    oldPassword,
    userId,
  }: {
    newPassword: string;
    oldPassword: string;
    userId: string;
  }) => changePasswordRequest({ newPassword, oldPassword, userId }),
);

export const $profile = createStore<null | Profile>(null)
  .on(fetchProfileFx.doneData, (_, profile) => profile)
  .on(updateProfileFx.doneData, (_, profile) => profile)
  .on(activateClientFx.doneData, (state, active) => {
    if (!state) {
      return state;
    }

    return {
      ...state,
      active_client: active.client,
      active_client_id: active.client.id,
    };
  })
  .reset(sessionEnded);

export const $clients = createStore<ProfileClientListItem[]>([])
  .on(fetchClientsFx.doneData, (_, clients) => clients)
  .on(activateClientFx.doneData, (state, active) =>
    state.map((item) => ({
      ...item,
      is_active: item.client.id === active.client.id,
    })),
  )
  .reset(sessionEnded);

export const $clientDetails = createStore<null | ProfileClient>(null)
  .on(fetchClientDetailsFx.doneData, (_, details) => details)
  .reset(sessionEnded);

export const $conditions = createStore<ClientConditions | null>(null)
  .on(fetchConditionsFx.doneData, (_, conditions) => conditions)
  .reset(sessionEnded);

export const $isProfilePending = combine(
  fetchProfileFx.pending,
  fetchClientsFx.pending,
  fetchClientDetailsFx.pending,
  fetchConditionsFx.pending,
  updateProfileFx.pending,
  activateClientFx.pending,
  changePasswordFx.pending,
  (...flags) => flags.some(Boolean),
);

export const $profileError = createStore<null | string>(null)
  .on(fetchProfileFx, () => null)
  .on(fetchProfileFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить профиль'),
  )
  .on(fetchClientsFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить список кабинетов'),
  )
  .on(fetchClientDetailsFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить реквизиты'),
  )
  .on(fetchConditionsFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить условия'),
  )
  .reset(sessionEnded);

/** Инкремент после успешной смены пароля — форма сбрасывает поля. */
export const $passwordChangeVersion = createStore(0)
  .on(changePasswordFx.done, (version) => version + 1)
  .reset(sessionEnded);

const $wantProfile = createStore(false)
  .on(cabinetProfileOpened, () => true)
  .reset(sessionEnded);

const $isProfileFetched = createStore(false)
  .on(fetchProfileFx, () => true)
  .on(fetchProfileFx.fail, () => false)
  .reset(sessionEnded);

const $isClientsFetched = createStore(false)
  .on(fetchClientsFx, () => true)
  .on(fetchClientsFx.fail, () => false)
  .reset(sessionEnded);

const $profileRequestUserId = combine(
  $wantProfile,
  $isProfileFetched,
  fetchProfileFx.pending,
  $userId,
  (want, fetched, pending, userId) =>
    want && !fetched && !pending && isUserId(userId) ? userId : null,
);

const $clientsRequestUserId = combine(
  $wantProfile,
  $isClientsFetched,
  fetchClientsFx.pending,
  $userId,
  (want, fetched, pending, userId) =>
    want && !fetched && !pending && isUserId(userId) ? userId : null,
);

/** clientID, для которого уже запросили details+conditions. */
const $clientExtrasKey = createStore<null | string>(null)
  .on(fetchClientDetailsFx, (_, { clientID }) => clientID)
  .on(fetchClientDetailsFx.fail, () => null)
  .on(fetchConditionsFx.fail, () => null)
  .reset(sessionEnded);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: [cabinetProfileOpened, $userId],
  source: $profileRequestUserId,
  filter: isUserId,
  target: fetchProfileFx,
});

sample({
  clock: [cabinetProfileOpened, $userId],
  source: $clientsRequestUserId,
  filter: isUserId,
  target: fetchClientsFx,
});

sample({
  clock: fetchProfileFx.doneData,
  source: {
    extrasKey: $clientExtrasKey,
    pendingDetails: fetchClientDetailsFx.pending,
    pendingConditions: fetchConditionsFx.pending,
    userId: $userId,
  },
  filter: ({ extrasKey, pendingConditions, pendingDetails, userId }, profile) =>
    isUserId(userId) &&
    profile.active_client_id.length > 0 &&
    extrasKey !== profile.active_client_id &&
    !pendingDetails &&
    !pendingConditions,
  fn: ({ userId }, profile) => ({
    clientID: profile.active_client_id,
    userId: userId as string,
  }),
  target: [fetchClientDetailsFx, fetchConditionsFx],
});

sample({
  clock: activateClientFx.doneData,
  source: $userId,
  filter: isUserId,
  fn: (userId, active) => ({
    clientID: active.client.id,
    userId,
  }),
  target: [fetchClientDetailsFx, fetchConditionsFx],
});

sample({
  clock: profileSaveRequested,
  source: $userId,
  filter: isUserId,
  fn: (userId, payload) => ({ payload, userId }),
  target: updateProfileFx,
});

sample({
  clock: clientActivateRequested,
  source: $userId,
  filter: isUserId,
  fn: (userId, clientID) => ({ clientID, userId }),
  target: activateClientFx,
});

sample({
  clock: passwordChangeRequested,
  source: $userId,
  filter: isUserId,
  fn: (userId, passwords) => ({ ...passwords, userId }),
  target: changePasswordFx,
});

sample({
  clock: updateProfileFx.done,
  fn: () => ({ message: 'Профиль сохранён', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: updateProfileFx.failData,
  fn: (error) => ({
    message: toDisplayErrorMessage(error, 'Не удалось обновить профиль'),
    tone: 'error' as const,
  }),
  target: toastShown,
});

sample({
  clock: activateClientFx.done,
  fn: () => ({ message: 'Активный кабинет переключён', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: activateClientFx.failData,
  fn: (error) => ({
    message: toDisplayErrorMessage(error, 'Не удалось переключить кабинет'),
    tone: 'error' as const,
  }),
  target: toastShown,
});

sample({
  clock: changePasswordFx.done,
  fn: () => ({ message: 'Пароль изменён', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: changePasswordFx.failData,
  fn: (error) => ({
    message: toDisplayErrorMessage(error, 'Не удалось сменить пароль'),
    tone: 'error' as const,
  }),
  target: toastShown,
});

/* eslint-enable perfectionist/sort-objects */
