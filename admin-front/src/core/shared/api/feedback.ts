import type { OrderFeedback } from '@/core/shared/server/portal/types';

import { portalRequest } from './portalClient';

export type { OrderFeedback } from '@/core/shared/server/portal/types';

export const ORDER_RATING_MAX = 5;

export type CreateFeedbackPayload = {
  clientName?: string;
  orderId: string;
  orderStatus?: string;
  rating: number;
  text?: string;
  userId?: string;
};

export const listOrderFeedback = async (params?: {
  orderId?: string;
  userId?: string;
}): Promise<OrderFeedback[]> => {
  const search = new URLSearchParams();

  if (params?.orderId) {
    search.set('orderId', params.orderId);
  }

  if (params?.userId) {
    search.set('userId', params.userId);
  }

  const query = search.toString();
  const result = await portalRequest<{ items: OrderFeedback[] }>({
    fallback: 'Не удалось загрузить отзывы',
    path: query ? `/feedback?${query}` : '/feedback',
  });

  return result.items;
};

export const createOrderFeedback = (payload: CreateFeedbackPayload): Promise<OrderFeedback> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось отправить отзыв',
    method: 'POST',
    path: '/feedback',
  });
