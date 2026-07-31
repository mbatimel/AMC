import type { OrderFeedback } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  createId,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type CreateFeedbackBody = {
  clientName?: string;
  orderId?: string;
  orderStatus?: string;
  rating?: number;
  text?: string;
  userId?: string;
};

export const GET = (request: Request): Response => {
  const params = new URL(request.url).searchParams;
  const userId = params.get('userId');
  const orderId = params.get('orderId');
  let items = readPortalState().feedback;

  if (userId) {
    items = items.filter((item) => item.user_id === userId);
  }

  if (orderId) {
    items = items.filter((item) => item.order_id === orderId);
  }

  return apiOk({ items });
};

export const POST = async (request: Request): Promise<Response> => {
  const body = await readJsonBody<CreateFeedbackBody>(request);
  const rating = Number(body?.rating);

  if (!body?.orderId || !Number.isFinite(rating) || rating < 1 || rating > 5) {
    return apiFail(400, 'Укажите заказ и оценку от 1 до 5');
  }

  const created: OrderFeedback = {
    client_name: body.clientName?.trim() ?? '',
    created_at: new Date().toISOString(),
    id: createId('feedback'),
    order_id: body.orderId,
    order_status: body.orderStatus ?? '',
    rating,
    text: body.text?.trim() ?? '',
    user_id: body.userId ?? '',
  };

  updatePortalState((draft) => {
    draft.feedback = [created, ...draft.feedback.filter((item) => item.order_id !== body.orderId)];
  });

  appendAuditEntry('Портал', `Получен отзыв по заказу ${created.order_id}: ${rating}/5`);

  return apiOk(created);
};
