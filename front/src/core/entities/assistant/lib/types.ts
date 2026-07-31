import type { ProductListItem } from '@/core/shared/api/products';

export type AssistantAuthor = 'bot' | 'operator' | 'system' | 'user';

export type AssistantMessage = {
  author: AssistantAuthor;
  id: string;
  offerOperator?: boolean;
  products?: ProductListItem[];
  text: string;
  time: string;
};
