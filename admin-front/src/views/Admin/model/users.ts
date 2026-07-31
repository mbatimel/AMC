import { createEffect, createEvent, createStore, sample } from 'effector';

import type { PortalUser } from '@/core/shared/api/portalUsers';
import type { SignupRequest, SignupRequestStatus } from '@/core/shared/api/signupRequests';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { invitePortalUser, listPortalUsers, patchPortalUser } from '@/core/shared/api/portalUsers';
import { decideSignupRequest, listSignupRequests } from '@/core/shared/api/signupRequests';
import { toastShown } from '@/core/shared/ui/Toast/model';

export const adminUsersOpened = createEvent();
export const portalUserInvited = createEvent<{ company: string; email: string }>();
export const portalUserToggled = createEvent<{ id: string; isActive: boolean }>();
export const portalUserPasswordReset = createEvent<string>();
export const signupDecided = createEvent<{
  id: string;
  rejectReason?: string;
  status: SignupRequestStatus;
}>();

export const fetchPortalUsersFx = createEffect(() => listPortalUsers());
export const fetchSignupRequestsFx = createEffect(() => listSignupRequests());

export const inviteUserFx = createEffect(async ({ company, email }: { company: string; email: string }) =>
  invitePortalUser({ company, email, role: 'admin' }),
);

export const toggleUserFx = createEffect(async ({ id, isActive }: { id: string; isActive: boolean }) =>
  patchPortalUser(id, { isActive }),
);

export const resetUserPasswordFx = createEffect(async (id: string) =>
  patchPortalUser(id, { passwordReset: true }),
);

export const decideSignupFx = createEffect(
  async ({
    id,
    rejectReason,
    status,
  }: {
    id: string;
    rejectReason?: string;
    status: SignupRequestStatus;
  }) => decideSignupRequest(id, { rejectReason, status }),
);

export const $portalUsers = createStore<PortalUser[]>([])
  .on(fetchPortalUsersFx.doneData, (_, users) => users)
  .on(inviteUserFx.doneData, (state, user) => [user, ...state])
  .on([toggleUserFx.doneData, resetUserPasswordFx.doneData], (state, user) =>
    state.map((item) => (item.id === user.id ? user : item)),
  );

export const $signupRequests = createStore<SignupRequest[]>([])
  .on(fetchSignupRequestsFx.doneData, (_, items) => items)
  .on(decideSignupFx.doneData, (state, updated) =>
    state.map((item) => (item.id === updated.id ? updated : item)),
  );

export const $isUsersPending = fetchPortalUsersFx.pending;

export const $usersError = createStore<null | string>(null)
  .on([fetchPortalUsersFx, fetchSignupRequestsFx], () => null)
  .on(
    [
      fetchPortalUsersFx.failData,
      fetchSignupRequestsFx.failData,
      inviteUserFx.failData,
      toggleUserFx.failData,
      decideSignupFx.failData,
    ],
    (_, error) => toDisplayErrorMessage(error, 'Не удалось выполнить операцию'),
  );

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> fn -> target */

sample({
  clock: adminUsersOpened,
  target: [fetchPortalUsersFx, fetchSignupRequestsFx],
});

sample({
  clock: portalUserInvited,
  target: inviteUserFx,
});

sample({
  clock: portalUserToggled,
  target: toggleUserFx,
});

sample({
  clock: portalUserPasswordReset,
  target: resetUserPasswordFx,
});

sample({
  clock: signupDecided,
  target: decideSignupFx,
});

sample({
  clock: decideSignupFx.done,
  target: fetchPortalUsersFx,
});

sample({
  clock: [inviteUserFx.done, toggleUserFx.done, decideSignupFx.done],
  fn: () => ({ message: 'Готово', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: resetUserPasswordFx.done,
  fn: () => ({ message: 'Ссылка для сброса пароля отправлена', tone: 'success' as const }),
  target: toastShown,
});

/* eslint-enable perfectionist/sort-objects */
