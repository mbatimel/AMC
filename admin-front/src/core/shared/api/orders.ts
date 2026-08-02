import {
  assertApiSuccess,
  fetchWithNetworkFallback,
  omitEmptyParams,
  parseApiErrorMessage,
} from './parseApiError';

export type ListOrdersResult = {
  items: Order[];
  total: number;
};

export type Order = {
  contact_name: string;
  created_at: string;
  delivery_address: string;
  delivery_type: string;
  id: string;
  items: OrderItem[];
  number: string;
  payment_status: string;
  status: string;
  total: number;
};

export type OrderItem = {
  id: string;
  price: number;
  product_name: string;
  qty: number;
  sku: string;
  total: number;
};

export class OrdersApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'OrdersApiError';
    this.status = status;
  }
}

const asNumber = (value: unknown, fallback = 0): number =>
  typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const asString = (value: unknown, fallback = ''): string =>
  typeof value === 'string' ? value : fallback;

const parseOrderItem = (value: unknown): null | OrderItem => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);

  if (!id) {
    return null;
  }

  return {
    id,
    price: asNumber(record.price),
    product_name: asString(record.product_name),
    qty: asNumber(record.qty),
    sku: asString(record.sku),
    total: asNumber(record.total),
  };
};

const parseOrder = (value: unknown): null | Order => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);

  if (!id) {
    return null;
  }

  const items = Array.isArray(record.items)
    ? record.items.flatMap((item) => {
        const parsed = parseOrderItem(item);

        return parsed ? [parsed] : [];
      })
    : [];

  return {
    contact_name: asString(record.contact_name),
    created_at: asString(record.created_at),
    delivery_address: asString(record.delivery_address),
    delivery_type: asString(record.delivery_type),
    id,
    items,
    number: asString(record.number),
    payment_status: asString(record.payment_status),
    status: asString(record.status),
    total: asNumber(record.total),
  };
};

/** История заказов конкретного пользователя. Бэк резолвит контрагента по X-User-Id. */
export const listUserOrdersRequest = async (userId: string): Promise<ListOrdersResult> => {
  const fallback = 'Не удалось загрузить историю заказов';
  const query = omitEmptyParams({ limit: 100 });
  const response = await fetchWithNetworkFallback(
    `/api/v1/orders${query}`,
    { headers: { 'X-User-Id': userId } },
    fallback,
  );

  if (!response.ok) {
    throw new OrdersApiError(response.status, await parseApiErrorMessage(response, fallback));
  }

  const record = assertApiSuccess(await response.json(), fallback);
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new OrdersApiError(500, 'Некорректный ответ заказов');
  }

  const data = payload as Record<string, unknown>;
  const rawItems = Array.isArray(data.items) ? data.items : [];
  const pagination =
    typeof data.pagination === 'object' && data.pagination !== null
      ? (data.pagination as Record<string, unknown>)
      : {};

  return {
    items: rawItems.flatMap((item) => {
      const parsed = parseOrder(item);

      return parsed ? [parsed] : [];
    }),
    total: asNumber(pagination.total),
  };
};
