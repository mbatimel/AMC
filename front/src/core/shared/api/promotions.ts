import { assertApiSuccess, omitEmptyParams, parseApiErrorMessage } from './parseApiError';

export type Promotion = {
  discount_percent: number;
  ends_at: string;
  id: string;
  name: string;
  products: PromotionProduct[];
  starts_at: string;
  status: PromotionStatus;
};

export type PromotionProduct = {
  min_qty: number;
  product_id: string;
};

export type PromotionStatus = 'active' | 'ended' | 'scheduled';

export class PromotionsApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'PromotionsApiError';
    this.status = status;
  }
}

const asNumber = (value: unknown, fallback = 0): number =>
  typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const asString = (value: unknown, fallback = ''): string =>
  typeof value === 'string' ? value : fallback;

const isPromotionStatus = (value: unknown): value is PromotionStatus =>
  value === 'active' || value === 'ended' || value === 'scheduled';

const parsePromotionProducts = (value: unknown): PromotionProduct[] => {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.flatMap((item) => {
    if (typeof item !== 'object' || item === null) {
      return [];
    }

    const record = item as Record<string, unknown>;
    const productId = asString(record.product_id);

    if (!productId) {
      return [];
    }

    return [{ min_qty: Math.max(1, asNumber(record.min_qty, 1)), product_id: productId }];
  });
};

const parsePromotion = (value: unknown): null | Promotion => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);
  const name = asString(record.name);

  if (!id || !name) {
    return null;
  }

  return {
    discount_percent: asNumber(record.discount_percent),
    ends_at: asString(record.ends_at),
    id,
    name,
    products: parsePromotionProducts(record.products),
    starts_at: asString(record.starts_at),
    status: isPromotionStatus(record.status) ? record.status : 'ended',
  };
};

export const listPromotionsRequest = async (): Promise<Promotion[]> => {
  const query = omitEmptyParams({ limit: 100, offset: 0 });
  const response = await fetch(`/api/v1/promotions${query}`);

  if (!response.ok) {
    throw new PromotionsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить акции'),
    );
  }

  const data: unknown = await response.json();
  const record = assertApiSuccess(data, 'Не удалось загрузить акции');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    return [];
  }

  const payloadRecord = payload as Record<string, unknown>;
  const rawItems = Array.isArray(payloadRecord.items) ? payloadRecord.items : [];

  return rawItems.flatMap((item) => {
    const parsed = parsePromotion(item);

    return parsed ? [parsed] : [];
  });
};

export const getPromotionRequest = async (promotionId: string): Promise<Promotion> => {
  const response = await fetch(`/api/v1/promotions/${promotionId}`);

  if (!response.ok) {
    throw new PromotionsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить акцию'),
    );
  }

  const data: unknown = await response.json();
  const record = assertApiSuccess(data, 'Не удалось загрузить акцию');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new PromotionsApiError(500, 'Некорректный ответ сервера');
  }

  const payloadRecord = payload as Record<string, unknown>;
  const promotionPayload = payloadRecord.promotion ?? payload;
  const promotion = parsePromotion(promotionPayload);

  if (!promotion) {
    throw new PromotionsApiError(500, 'Некорректный ответ сервера');
  }

  return promotion;
};
