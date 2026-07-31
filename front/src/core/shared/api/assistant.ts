import type { ProductListItem } from './products';

import { portalRequest } from './portalClient';

export type AssistantAskPayload = {
  history: { author: string; text: string }[];
  message: string;
};

export type AssistantReply = {
  /** false — ИИ не подключён или недоступен, фронт использует встроенную логику. */
  configured: boolean;
  offerOperator?: boolean;
  products?: ProductListItem[];
  reply?: string;
};

export const askAssistantRequest = (payload: AssistantAskPayload): Promise<AssistantReply> =>
  portalRequest({
    body: payload,
    fallback: 'Помощник временно недоступен',
    method: 'POST',
    path: '/assistant',
  });
