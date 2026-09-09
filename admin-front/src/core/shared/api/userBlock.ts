/** Поля при блокировке пользователя (все необязательные). */
export type UserBlockPayload = {
  contactEmail?: string;
  contactName?: string;
  contactPhone?: string;
  reason?: string;
};
