import {
  assertApiSuccess,
  fetchWithNetworkFallback,
  omitEmptyParams,
  parseApiErrorMessage,
} from './parseApiError';

export type CityItem = {
  city: string;
  id: string;
};

export type ListOrdersParams = {
  /** Пустой — бэк резолвит контрагента по X-User-Id. */
  clientID?: string;
  limit?: number;
  offset?: number;
  paymentStatus?: string;
  sort?: string;
  status?: string;
  userId: string;
};

export type ListOrdersResult = {
  items: Order[];
  pagination: OrdersPagination;
};

export type Order = {
  client_id: string;
  comment: string;
  contact_name: string;
  created_at: string;
  delivery_address: string;
  delivery_type: string;
  discount_total: number;
  documents: OrderDocument[];
  email: string;
  history: OrderHistoryItem[];
  id: string;
  items: OrderItem[];
  number: string;
  payment_status: string;
  phone: string;
  status: string;
  subtotal: number;
  total: number;
  updated_at: string;
  user_id: string;
  vat: number;
};

export type OrderDocument = {
  created_at: string;
  id: string;
  name: string;
  order_id: string;
  type: string;
  url: string;
};

export type OrderHistoryItem = {
  changed_by: string;
  comment: string;
  created_at: string;
  id: string;
  order_id: string;
  payment_status: string;
  status: string;
};

export type OrderItem = {
  id: string;
  order_id: string;
  price: number;
  product_id: string;
  product_name: string;
  qty: number;
  sku: string;
  total: number;
  vat: number;
};

export type OrdersPagination = {
  limit: number;
  offset: number;
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

const parseCities = (data: unknown): CityItem[] => {
  if (typeof data !== 'object' || data === null) {
    throw new OrdersApiError(500, 'Некорректный ответ сервера');
  }

  const record = data as Record<string, unknown>;

  if (record.error === true) {
    const errorText = record.errorText ?? record.ErrorText;

    throw new OrdersApiError(
      500,
      typeof errorText === 'string' && errorText.length > 0
        ? errorText
        : 'Не удалось загрузить список городов',
    );
  }

  const cities = record.data;

  if (!Array.isArray(cities)) {
    throw new OrdersApiError(500, 'Некорректный ответ сервера');
  }

  return cities.flatMap((item) => {
    if (typeof item !== 'object' || item === null) {
      return [];
    }

    const cityRecord = item as Record<string, unknown>;
    const id = cityRecord.id;
    const city = cityRecord.city;

    if (
      typeof id !== 'string' ||
      id.length === 0 ||
      typeof city !== 'string' ||
      city.length === 0
    ) {
      return [];
    }

    return [{ city, id }];
  });
};

const parseErrorMessage = async (response: Response): Promise<string> => {
  try {
    const data: unknown = await response.json();

    if (typeof data === 'object' && data !== null) {
      const record = data as Record<string, unknown>;
      const errorText = record.errorText ?? record.ErrorText;

      if (typeof errorText === 'string' && errorText.length > 0) {
        return errorText;
      }
    }
  } catch {
    // ignore JSON parse errors
  }

  if (response.status === 502 || response.status === 503 || response.status === 504) {
    return 'Сервис временно недоступен. Попробуйте позже';
  }

  return 'Не удалось загрузить список городов';
};

export const getCitiesRequest = async (): Promise<CityItem[]> => {
  const response = await fetchWithNetworkFallback(
    '/api/v1/orders/cities',
    undefined,
    'Не удалось загрузить список городов',
  );

  if (!response.ok) {
    throw new OrdersApiError(response.status, await parseErrorMessage(response));
  }

  const data: unknown = await response.json();
  const cities = parseCities(data);

  if (cities.length === 0) {
    throw new OrdersApiError(500, 'Список городов пуст');
  }

  return cities;
};

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
    order_id: asString(record.order_id),
    price: asNumber(record.price),
    product_id: asString(record.product_id),
    product_name: asString(record.product_name),
    qty: asNumber(record.qty),
    sku: asString(record.sku),
    total: asNumber(record.total),
    vat: asNumber(record.vat),
  };
};

const parseOrderDocument = (value: unknown): null | OrderDocument => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);

  if (!id) {
    return null;
  }

  return {
    created_at: asString(record.created_at),
    id,
    name: asString(record.name),
    order_id: asString(record.order_id),
    type: asString(record.type),
    url: asString(record.url),
  };
};

const parseOrderHistoryItem = (value: unknown): null | OrderHistoryItem => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);

  if (!id) {
    return null;
  }

  return {
    changed_by: asString(record.changed_by),
    comment: asString(record.comment),
    created_at: asString(record.created_at),
    id,
    order_id: asString(record.order_id),
    payment_status: asString(record.payment_status),
    status: asString(record.status),
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

  const documents = Array.isArray(record.documents)
    ? record.documents.flatMap((doc) => {
        const parsed = parseOrderDocument(doc);

        return parsed ? [parsed] : [];
      })
    : [];

  const history = Array.isArray(record.history)
    ? record.history.flatMap((item) => {
        const parsed = parseOrderHistoryItem(item);

        return parsed ? [parsed] : [];
      })
    : [];

  return {
    client_id: asString(record.client_id),
    comment: asString(record.comment),
    contact_name: asString(record.contact_name),
    created_at: asString(record.created_at),
    delivery_address: asString(record.delivery_address),
    delivery_type: asString(record.delivery_type),
    discount_total: asNumber(record.discount_total),
    documents,
    email: asString(record.email),
    history,
    id,
    items,
    number: asString(record.number),
    payment_status: asString(record.payment_status),
    phone: asString(record.phone),
    status: asString(record.status),
    subtotal: asNumber(record.subtotal),
    total: asNumber(record.total),
    updated_at: asString(record.updated_at),
    user_id: asString(record.user_id),
    vat: asNumber(record.vat),
  };
};

export const listOrdersRequest = async ({
  userId,
  ...params
}: ListOrdersParams): Promise<ListOrdersResult> => {
  const query = omitEmptyParams({
    clientID: params.clientID,
    limit: params.limit,
    offset: params.offset,
    paymentStatus: params.paymentStatus,
    sort: params.sort,
    status: params.status,
  });
  // Со слэшем: nginx location `/api/v1/orders/`; без слэша — 301 на http://… → CORS
  // (браузер ещё и кэширует 301 на GET).
  const response = await fetchWithNetworkFallback(
    `/api/v1/orders/${query}`,
    {
      headers: { 'X-User-Id': userId },
    },
    'Не удалось загрузить заказы',
  );

  if (!response.ok) {
    throw new OrdersApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить заказы'),
    );
  }

  const record = assertApiSuccess(await response.json(), 'Некорректный ответ заказов');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new OrdersApiError(500, 'Некорректный ответ заказов');
  }

  const data = payload as Record<string, unknown>;
  const rawItems = Array.isArray(data.items) ? data.items : [];
  const paginationRaw =
    typeof data.pagination === 'object' && data.pagination !== null
      ? (data.pagination as Record<string, unknown>)
      : {};

  return {
    items: rawItems.flatMap((item) => {
      const parsed = parseOrder(item);

      return parsed ? [parsed] : [];
    }),
    pagination: {
      limit: asNumber(paginationRaw.limit, params.limit ?? 20),
      offset: asNumber(paginationRaw.offset, params.offset ?? 0),
      total: asNumber(paginationRaw.total),
    },
  };
};

export type CreateOrderPayload = {
  /** Пустой — бэк резолвит контрагента по X-User-Id. */
  clientID?: string;
  comment?: string;
  contactName: string;
  deliveryAddress: string;
  deliveryType: string;
  email?: string;
  phone?: string;
  userId: string;
};

export const createOrderRequest = async ({
  userId,
  ...params
}: CreateOrderPayload): Promise<Order> => {
  const query = omitEmptyParams({
    clientID: params.clientID,
    comment: params.comment,
    contactName: params.contactName,
    deliveryAddress: params.deliveryAddress,
    deliveryType: params.deliveryType,
    email: params.email,
    phone: params.phone,
  });
  // Со слэшем: nginx location `/api/v1/orders/`; без слэша — 301 на http://… → CORS.
  const response = await fetchWithNetworkFallback(
    `/api/v1/orders/${query}`,
    {
      headers: { 'X-User-Id': userId },
      method: 'POST',
    },
    'Не удалось оформить заказ',
  );

  if (!response.ok) {
    throw new OrdersApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось оформить заказ'),
    );
  }

  const record = assertApiSuccess(await response.json(), 'Некорректный ответ создания заказа');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new OrdersApiError(500, 'Некорректный ответ создания заказа');
  }

  const data = payload as Record<string, unknown>;
  const order = parseOrder(data.order ?? data);

  if (!order) {
    throw new OrdersApiError(500, 'Некорректный ответ создания заказа');
  }

  return order;
};

export type GetOrderParams = {
  /** Пустой — бэк резолвит контрагента по X-User-Id. */
  clientID?: string;
  orderID: string;
  userId: string;
};

export const getOrderRequest = async ({
  clientID,
  orderID,
  userId,
}: GetOrderParams): Promise<Order> => {
  const query = omitEmptyParams({ clientID, orderID });
  const response = await fetchWithNetworkFallback(
    `/api/v1/orders/id${query}`,
    {
      headers: { 'X-User-Id': userId },
    },
    'Не удалось загрузить заказ',
  );

  if (!response.ok) {
    throw new OrdersApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить заказ'),
    );
  }

  const record = assertApiSuccess(await response.json(), 'Некорректный ответ заказа');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new OrdersApiError(500, 'Некорректный ответ заказа');
  }

  const data = payload as Record<string, unknown>;
  const order = parseOrder(data.order ?? data);

  if (!order) {
    throw new OrdersApiError(500, 'Некорректный ответ заказа');
  }

  return order;
};
