import type {
  SupportRequest,
  SupportRequestStatus,
} from '@/core/shared/server/portal/types';

import { portalRequest } from './portalClient';

export type { SupportRequest, SupportRequestStatus } from '@/core/shared/server/portal/types';

export type CreateSupportRequestPayload = {
  clientName?: string;
  contact?: string;
  orderId?: string;
  severity: number;
  source?: 'assistant' | 'form';
  subject: string;
  text: string;
  userId?: string;
};

export const SUPPORT_SEVERITY_OPTIONS = [
  { label: '1 — вопрос или пожелание', value: 1 },
  { label: '2 — небольшая проблема', value: 2 },
  { label: '3 — влияет на сценарий работы', value: 3 },
  { label: '4 — серьёзная проблема', value: 4 },
  { label: '5 — критическая ошибка', value: 5 },
];

export const SUPPORT_STATUS_LABELS: Record<SupportRequestStatus, string> = {
  closed: 'Закрыто',
  in_progress: 'В работе',
  new: 'Новое',
};

export const createSupportRequest = (
  payload: CreateSupportRequestPayload,
): Promise<SupportRequest> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось отправить обращение',
    method: 'POST',
    path: '/support',
  });

export const listSupportRequests = async (params?: {
  status?: SupportRequestStatus;
  userId?: string;
}): Promise<SupportRequest[]> => {
  const search = new URLSearchParams();

  if (params?.status) {
    search.set('status', params.status);
  }

  if (params?.userId) {
    search.set('userId', params.userId);
  }

  const query = search.toString();
  const result = await portalRequest<{ items: SupportRequest[] }>({
    fallback: 'Не удалось загрузить обращения',
    path: query ? `/support?${query}` : '/support',
  });

  return result.items;
};

export const patchSupportRequest = (
  id: string,
  payload: { answer?: string; status?: SupportRequestStatus },
): Promise<SupportRequest> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось обновить обращение',
    method: 'PATCH',
    path: `/support/${id}`,
  });
