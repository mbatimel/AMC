import { createEffect, createEvent, createStore, sample } from 'effector';

import type { PortalUser } from '@/core/shared/api/portalUsers';
import type { SignupRequest, SignupRequestStatus } from '@/core/shared/api/signupRequests';
import type { RealUser } from '@/core/shared/api/users';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { invitePortalUser, patchPortalUser } from '@/core/shared/api/portalUsers';
import {
  decideSignupRequest,
  listSignupRequests,
  rejectSignupRequest,
} from '@/core/shared/api/signupRequests';
import { listUsersRequest, setUserActiveRequest } from '@/core/shared/api/users';
import { toastShown } from '@/core/shared/ui/Toast/model';

/** Приводим пользователя реального сервиса `users` к виду, который уже умеет рисовать UI. */
export const toPortalUser = (user: RealUser): PortalUser => ({
  company: user.company_name,
  contact: [user.last_name, user.first_name, user.middle_name].filter(Boolean).join(' '),
  created_at: user.created_at,
  email: user.email,
  id: user.id,
  inn: user.inn,
  is_active: user.is_active,
  phone: user.phone,
  role: user.role === 'admin' ? 'admin' : 'client',
});

export const adminUsersOpened = createEvent();
export const portalUserInvited = createEvent<{ company: string; email: string }>();
export const portalUserToggled = createEvent<{ id: string; isActive: boolean }>();
export const portalUserPasswordReset = createEvent<string>();
export const signupDecided = createEvent<{
  id: string;
  rejectReason?: string;
  status: SignupRequestStatus;
}>();
export const signupRejected = createEvent<{ id: string; reason: string }>();

export const fetchPortalUsersFx = createEffect(async () => {
  const { items } = await listUsersRequest();

  return items.map(toPortalUser);
});
export const fetchSignupRequestsFx = createEffect(() => listSignupRequests());

export const inviteUserFx = createEffect(
  async ({ company, email }: { company: string; email: string }) =>
    invitePortalUser({ company, email, role: 'admin' }),
);

export const toggleUserFx = createEffect(
  async ({ id, isActive }: { id: string; isActive: boolean }) => {
    const user = await setUserActiveRequest(id, isActive);

    return toPortalUser(user);
  },
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

export const rejectSignupFx = createEffect(({ id, reason }: { id: string; reason: string }) =>
  rejectSignupRequest(id, reason),
);

export const $portalUsers = createStore<PortalUser[]>([])
  .on(fetchPortalUsersFx.doneData, (_, users) => users)
  .on(inviteUserFx.doneData, (state, user) => [user, ...state])
  .on([toggleUserFx.doneData, resetUserPasswordFx.doneData], (state, user) =>
    state.map((item) => (item.id === user.id ? user : item)),
  );

export const $signupRequests = createStore<SignupRequest[]>([])
  .on(fetchSignupRequestsFx.doneData, (_, items) => items)
  .on([decideSignupFx.doneData, rejectSignupFx.doneData], (state, updated) =>
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
      rejectSignupFx.failData,
    ],
    (_, error) => toDisplayErrorMessage(error, 'Не удалось выполнить операцию'),
  );

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
  clock: signupRejected,
  target: rejectSignupFx,
});

sample({
  clock: [decideSignupFx.done, rejectSignupFx.done],
  target: fetchPortalUsersFx,
});

sample({
  clock: [inviteUserFx.done, toggleUserFx.done, decideSignupFx.done, rejectSignupFx.done],
  fn: () => ({ message: 'Готово', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: resetUserPasswordFx.done,
  fn: () => ({ message: 'Ссылка для сброса пароля отправлена', tone: 'success' as const }),
  target: toastShown,
});
