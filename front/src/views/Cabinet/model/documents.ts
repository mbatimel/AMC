import { combine } from 'effector';

import type { OrderDocument } from '@/core/shared/api/orders';

import { $orders } from './orders';

export type CabinetDocument = OrderDocument & {
  orderNumber: string;
  orderStatus: string;
};

const DOCUMENT_TYPE_LABELS: Record<string, string> = {
  act: 'Акт',
  invoice: 'Счёт на оплату',
  torg12: 'ТОРГ-12',
  upd: 'УПД',
};

export const formatDocumentType = (type: string): string =>
  DOCUMENT_TYPE_LABELS[type] ?? (type || 'Документ');

/** Все закрывающие документы клиента — плоским списком по всем заказам. */
export const $cabinetDocuments = combine($orders, (orders): CabinetDocument[] =>
  orders
    .flatMap((order) =>
      (order.documents ?? []).map((document) => ({
        ...document,
        orderNumber: order.number || order.id,
        orderStatus: order.status,
      })),
    )
    .sort((a, b) => (b.created_at ?? '').localeCompare(a.created_at ?? '')),
);
