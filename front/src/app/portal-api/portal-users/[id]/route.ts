import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import { appendAuditEntry, readPortalState, updatePortalState } from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type RouteContext = { params: Promise<{ id: string }> };

type PatchPortalUserBody = {
  isActive?: boolean;
  passwordReset?: boolean;
};

export const PATCH = async (request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  const payload = await readJsonBody<PatchPortalUserBody>(request);

  if (!payload) {
    return apiFail(400, 'Некорректное тело запроса');
  }

  let found = false;
  let email = '';

  updatePortalState((draft) => {
    const user = draft.portal_users.find((item) => item.id === id);

    if (!user) {
      return;
    }

    found = true;
    email = user.email;

    if (typeof payload.isActive === 'boolean') {
      user.is_active = payload.isActive;
    }
  });

  if (!found) {
    return apiFail(404, 'Пользователь не найден');
  }

  appendAuditEntry(
    'Админ портала',
    payload.passwordReset
      ? `Сброшен пароль пользователя ${email}`
      : `${payload.isActive ? 'Разблокирован' : 'Заблокирован'} пользователь ${email}`,
  );

  return apiOk(readPortalState().portal_users.find((item) => item.id === id));
};
