import { createEffect, createEvent, createStore, sample } from 'effector';

import type { InvitePortalAdminResult, PortalUser } from '@/core/shared/api/portalUsers';
import type { UserBlockPayload } from '@/core/shared/api/userBlock';
import type { RealUser } from '@/core/shared/api/users';

import { $adminUserId } from '@/core/entities/adminSession';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { invitePortalUser } from '@/core/shared/api/portalUsers';
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

const isAdminUserId = (userId: null | string): userId is string => Boolean(userId);

export type ToggleUserPayload = {
  deactivate?: UserBlockPayload;
  id: string;
  isActive: boolean;
};

export const adminUsersOpened = createEvent();
export const portalUserInvited = createEvent<{ email: string; name: string }>();
export const portalUserToggled = createEvent<ToggleUserPayload>();
export const inviteResultDismissed = createEvent();

export const fetchPortalUsersFx = createEffect(async () => {
  const items = await listAllUsersRequest();

  return items.map(toPortalUser);
});

export const inviteUserFx = createEffect(
  ({
    email,
    name,
    userId,
  }: {
    email: string;
    name: string;
    userId: string;
  }): Promise<InvitePortalAdminResult> => invitePortalUser(userId, { email, name }),
);

export const toggleUserFx = createEffect<ToggleUserPayload, PortalUser, Error>(async (payload) => {
  const user = await setUserActiveRequest({
    deactivate: payload.deactivate,
    isActive: payload.isActive,
    userId: payload.id,
  });

  return toPortalUser(user);
});

export const $portalUsers = createStore<PortalUser[]>([])
  .on(fetchPortalUsersFx.doneData, (_, users) => users)
  .on(toggleUserFx.doneData, (state, user) =>
    state.map((item) => (item.id === user.id ? user : item)),
  );

export const $inviteResult = createStore<InvitePortalAdminResult | null>(null)
  .on(inviteUserFx.doneData, (_, result) => result)
  .reset([inviteResultDismissed, adminUsersOpened]);

export const $isUsersPending = fetchPortalUsersFx.pending;
export const $isInvitePending = inviteUserFx.pending;
export const $isTogglePending = toggleUserFx.pending;

export const $usersError = createStore<null | string>(null)
  .on([fetchPortalUsersFx, inviteUserFx, toggleUserFx], () => null)
  .on([fetchPortalUsersFx.failData, inviteUserFx.failData, toggleUserFx.failData], (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось выполнить операцию'),
  );

sample({
  clock: adminUsersOpened,
  target: fetchPortalUsersFx,
});

/* eslint-disable perfectionist/sort-objects -- effector sample option order */
sample({
  clock: portalUserInvited,
  source: $adminUserId,
  filter: isAdminUserId,
  fn: (userId, payload) => ({ ...payload, userId }),
  target: inviteUserFx,
});
/* eslint-enable perfectionist/sort-objects */

sample({
  clock: portalUserToggled,
  target: toggleUserFx,
});

sample({
  clock: inviteUserFx.done,
  target: fetchPortalUsersFx,
});

sample({
  clock: toggleUserFx.done,
  fn: () => ({ message: 'Готово', tone: 'success' as const }),
  target: toastShown,
});
