import { assertApiSuccess, omitEmptyParams, parseApiErrorMessage } from './parseApiError';

export type Cart = {
  client_id: string;
  discount_total: number;
  id: string;
  items: CartItem[];
  subtotal: number;
  total: number;
  user_id: string;
  vat: number;
};

export type CartItem = {
  cart_id: string;
  id: string;
  price: number;
  product_id: string;
  product_name: string;
  qty: number;
  sku: string;
  total: number;
  vat?: number;
};

type UserScopedParams = {
  clientID?: string;
  userId: string;
};

export class CartApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'CartApiError';
    this.status = status;
  }
}

const asNumber = (value: unknown, fallback = 0): number =>
  typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const asString = (value: unknown, fallback = ''): string =>
  typeof value === 'string' ? value : fallback;

const parseCartItem = (value: unknown): CartItem | null => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);
  const productId = asString(record.product_id);
  const productName = asString(record.product_name);
  const sku = asString(record.sku);

  if (!id || !productId) {
    return null;
  }

  const qty = asNumber(record.qty);
  const price = asNumber(record.price);

  return {
    cart_id: asString(record.cart_id),
    id,
    price,
    product_id: productId,
    product_name: productName || sku,
    qty,
    sku,
    total: asNumber(record.total, qty * price),
    vat: asNumber(record.vat),
  };
};

const parseCart = (data: unknown): Cart => {
  const record = assertApiSuccess(data, 'Некорректный ответ корзины');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new CartApiError(500, 'Некорректный ответ корзины');
  }

  const payloadRecord = payload as Record<string, unknown>;
  const cartPayload = payloadRecord.cart ?? payload;

  if (typeof cartPayload !== 'object' || cartPayload === null) {
    throw new CartApiError(500, 'Некорректный ответ корзины');
  }

  const cartRecord = cartPayload as Record<string, unknown>;
  const rawItems = Array.isArray(cartRecord.items) ? cartRecord.items : [];
  const items = rawItems.flatMap((item) => {
    const parsed = parseCartItem(item);

    return parsed ? [parsed] : [];
  });

  return {
    client_id: asString(cartRecord.client_id),
    discount_total: asNumber(cartRecord.discount_total),
    id: asString(cartRecord.id),
    items,
    subtotal: asNumber(cartRecord.subtotal),
    total: asNumber(cartRecord.total),
    user_id: asString(cartRecord.user_id),
    vat: asNumber(cartRecord.vat),
  };
};

const withUserHeaders = (userId: string): HeadersInit => ({
  'X-User-Id': userId,
});

const throwIfNotOk = async (response: Response, fallback: string): Promise<void> => {
  if (!response.ok) {
    throw new CartApiError(response.status, await parseApiErrorMessage(response, fallback));
  }
};

export const getCartRequest = async ({ clientID, userId }: UserScopedParams): Promise<Cart> => {
  const query = omitEmptyParams({ clientID });
  const response = await fetch(`/api/v1/cart${query}`, {
    headers: withUserHeaders(userId),
  });

  await throwIfNotOk(response, 'Не удалось загрузить корзину');

  return parseCart(await response.json());
};

export const addCartItemRequest = async ({
  clientID,
  productID,
  qty,
  userId,
}: UserScopedParams & { productID: string; qty: number }): Promise<Cart> => {
  const query = omitEmptyParams({ clientID, productID, qty });
  const response = await fetch(`/api/v1/cart/items${query}`, {
    headers: withUserHeaders(userId),
    method: 'POST',
  });

  await throwIfNotOk(response, 'Не удалось добавить товар в корзину');

  return parseCart(await response.json());
};

export const updateCartItemRequest = async ({
  cartItemID,
  clientID,
  qty,
  userId,
}: UserScopedParams & { cartItemID: string; qty: number }): Promise<Cart> => {
  const query = omitEmptyParams({ cartItemID, clientID, qty });
  const response = await fetch(`/api/v1/cart/items/${cartItemID}${query}`, {
    headers: withUserHeaders(userId),
    method: 'PATCH',
  });

  await throwIfNotOk(response, 'Не удалось обновить позицию корзины');

  return parseCart(await response.json());
};

export const deleteCartItemRequest = async ({
  cartItemID,
  clientID,
  userId,
}: UserScopedParams & { cartItemID: string }): Promise<Cart> => {
  const query = omitEmptyParams({ cartItemID, clientID });
  const response = await fetch(`/api/v1/cart/items/${cartItemID}${query}`, {
    headers: withUserHeaders(userId),
    method: 'DELETE',
  });

  await throwIfNotOk(response, 'Не удалось удалить позицию');

  return parseCart(await response.json());
};

export const clearCartRequest = async ({ clientID, userId }: UserScopedParams): Promise<Cart> => {
  const query = omitEmptyParams({ clientID });
  const response = await fetch(`/api/v1/cart${query}`, {
    headers: withUserHeaders(userId),
    method: 'DELETE',
  });

  await throwIfNotOk(response, 'Не удалось очистить корзину');

  return parseCart(await response.json());
};
