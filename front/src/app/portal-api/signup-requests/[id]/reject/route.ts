import { sendMail } from '@/core/shared/server/portal/mailer';
import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type RejectSignupBody = { reason?: string };

type RouteContext = { params: Promise<{ id: string }> };

/**
 * Отклонение заявки на регистрацию: помечает заявку отклонённой, удаляет
 * созданный по ней аккаунт (если он уже существует) и уведомляет заявителя
 * письмом с причиной отказа. Действие необратимо для уже решённых заявок.
 */
export const POST = async (request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  const payload = await readJsonBody<RejectSignupBody>(request);
  const reason = payload?.reason?.trim() ?? '';

  let found = false;
  let alreadyDecided = false;
  let removedUser = false;
  let requestEmail = '';

  updatePortalState((draft) => {
    const item = draft.signup_requests.find((request_) => request_.id === id);

    if (!item) {
      return;
    }

    found = true;

    if (item.status !== 'pending') {
      alreadyDecided = true;

      return;
    }

    item.status = 'rejected';
    item.reject_reason = reason;
    requestEmail = item.email;

    const userIndex = draft.portal_users.findIndex((user) => user.email === item.email);

    if (userIndex !== -1) {
      draft.portal_users.splice(userIndex, 1);
      removedUser = true;
    }
  });

  if (!found) {
    return apiFail(404, 'Заявка не найдена');
  }

  if (alreadyDecided) {
    return apiFail(409, 'Заявка уже обработана');
  }

  sendMail({
    subject: 'Заявка на регистрацию отклонена',
    text: reason
      ? `Ваша заявка на регистрацию отклонена. Причина: ${reason}`
      : 'Ваша заявка на регистрацию отклонена.',
    to: requestEmail,
  });

  appendAuditEntry(
    'Админ портала',
    `Отклонена заявка ${id}${removedUser ? ', аккаунт удалён' : ''}, уведомление отправлено на ${requestEmail}`,
  );

  return apiOk(readPortalState().signup_requests.find((item) => item.id === id));
};
