import type { SignupRequest, SignupRequestStatus } from '@/core/shared/server/portal/types';

import { portalRequest } from './portalClient';

export type { SignupRequest, SignupRequestStatus } from '@/core/shared/server/portal/types';

export const SIGNUP_STATUS_LABELS: Record<SignupRequestStatus, string> = {
  approved: 'Одобрена',
  pending: 'Ожидает',
  rejected: 'Отклонена',
};

export type CreateSignupRequestPayload = {
  company?: string;
  contact?: string;
  email: string;
  inn?: string;
  phone?: string;
  type: 'individual' | 'organization';
};

export const createSignupRequest = (payload: CreateSignupRequestPayload): Promise<SignupRequest> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось отправить заявку',
    method: 'POST',
    path: '/signup-requests',
  });

export const listSignupRequests = async (): Promise<SignupRequest[]> => {
  const result = await portalRequest<{ items: SignupRequest[] }>({
    fallback: 'Не удалось загрузить заявки',
    path: '/signup-requests',
  });

  return result.items;
};

export const decideSignupRequest = (
  id: string,
  payload: { rejectReason?: string; status: SignupRequestStatus },
): Promise<SignupRequest> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось обновить заявку',
    method: 'PATCH',
    path: `/signup-requests/${id}`,
  });
