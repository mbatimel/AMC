import type { EmailLogEntry } from './types';

import { createId, updatePortalState } from './store';

/**
 * Временная имитация почтового шлюза (см. комментарий в `store.ts`): письма
 * не уходят наружу, а складываются в portal-состояние и лог сервера. Когда
 * появится реальный сервис уведомлений — заменить на настоящую отправку.
 */
export const sendMail = ({
  subject,
  text,
  to,
}: {
  subject: string;
  text: string;
  to: string;
}): void => {
  const entry: EmailLogEntry = {
    created_at: new Date().toISOString(),
    id: createId('mail'),
    subject,
    text,
    to,
  };

  updatePortalState((state) => {
    state.email_log.unshift(entry);
    state.email_log = state.email_log.slice(0, 500);
  });

  console.info(`[portal-mail] -> ${to}: ${subject}`);
};
