import { createEffect, createEvent, createStore, sample } from 'effector';

import type { PortalUser } from '@/core/shared/api/portalUsers';
import type { RealUser } from '@/core/shared/api/users';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { invitePortalUser, patchPortalUser } from '@/core/shared/api/portalUsers';
import { listAllUsersRequest, setUserActiveRequest } from '@/core/shared/api/users';
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

export const fetchPortalUsersFx = createEffect(async () => {
  const items = await listAllUsersRequest();

  return items.map(toPortalUser);
});

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

export const $portalUsers = createStore<PortalUser[]>([])
  .on(fetchPortalUsersFx.doneData, (_, users) => users)
  .on(inviteUserFx.doneData, (state, user) => [user, ...state])
  .on([toggleUserFx.doneData, resetUserPasswordFx.doneData], (state, user) =>
    state.map((item) => (item.id === user.id ? user : item)),
  );

export const $isUsersPending = fetchPortalUsersFx.pending;

export const $usersError = createStore<null | string>(null)
  .on(fetchPortalUsersFx, () => null)
  .on([fetchPortalUsersFx.failData, inviteUserFx.failData, toggleUserFx.failData], (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось выполнить операцию'),
  );

sample({
  clock: adminUsersOpened,
  target: fetchPortalUsersFx,
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
  clock: [inviteUserFx.done, toggleUserFx.done],
  fn: () => ({ message: 'Готово', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: resetUserPasswordFx.done,
  fn: () => ({ message: 'Ссылка для сброса пароля отправлена', tone: 'success' as const }),
  target: toastShown,
});
