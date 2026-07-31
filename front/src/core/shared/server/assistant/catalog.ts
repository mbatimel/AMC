import type { ProductListItem } from '@/core/shared/api/products';

import { readApiBaseUrl } from './config';

export type CatalogSearchParams = {
  gost?: string;
  inStock?: boolean;
  limit?: number;
  material?: string;
  q?: string;
  size?: string;
};

const asNumber = (value: unknown, fallback = 0): number =>
  typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const asString = (value: unknown, fallback = ''): string =>
  typeof value === 'string' ? value : fallback;

const parseItem = (value: unknown): null | ProductListItem => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);
  const sku = asString(record.sku);
  const name = asString(record.name);

  if (!id || !sku || !name) {
    return null;
  }

  const images = Array.isArray(record.images)
    ? record.images.flatMap((image) => {
        if (typeof image !== 'object' || image === null) {
          return [];
        }

        const imageRecord = image as Record<string, unknown>;
        const url = asString(imageRecord.url);

        return url ? [{ id: asString(imageRecord.id, url), url }] : [];
      })
    : [];

  return {
    base_price: asNumber(record.base_price),
    brand_id: asString(record.brand_id) || undefined,
    brand_name: asString(record.brand_name) || undefined,
    category_id: asString(record.category_id) || undefined,
    category_name: asString(record.category_name) || undefined,
    client_price: asNumber(record.client_price),
    discount_percent: asNumber(record.discount_percent),
    gost: asString(record.gost) || undefined,
    id,
    images,
    is_published: record.is_published !== false,
    material: asString(record.material) || undefined,
    name,
    package_qty: Math.max(1, asNumber(record.package_qty, 1)),
    size: asString(record.size) || undefined,
    sku,
    stock_qty: asNumber(record.stock_qty),
  };
};

/**
 * Серверный поиск по каталогу (M-01). В отличие от клиентского клиента
 * ходит по абсолютному адресу backend, а не через rewrite Next.js.
 */
export const searchCatalogProducts = async (
  params: CatalogSearchParams,
): Promise<ProductListItem[]> => {
  const search = new URLSearchParams();

  search.set('limit', String(params.limit ?? 5));
  search.set('offset', '0');

  if (params.q) {
    search.set('q', params.q);
  }

  if (params.gost) {
    search.set('gost', params.gost);
  }

  if (params.material) {
    search.set('material', params.material);
  }

  if (params.size) {
    search.set('size', params.size);
  }

  if (params.inStock) {
    search.set('inStock', 'true');
  }

  try {
    const response = await fetch(`${readApiBaseUrl()}/api/v1/products?${search.toString()}`, {
      cache: 'no-store',
      signal: AbortSignal.timeout(10_000),
    });

    if (!response.ok) {
      return [];
    }

    const payload: unknown = await response.json();

    if (typeof payload !== 'object' || payload === null) {
      return [];
    }

    const data = (payload as Record<string, unknown>).data;

    if (typeof data !== 'object' || data === null) {
      return [];
    }

    const rawItems = (data as Record<string, unknown>).items;

    if (!Array.isArray(rawItems)) {
      return [];
    }

    return rawItems.flatMap((item) => {
      const parsed = parseItem(item);

      return parsed ? [parsed] : [];
    });
  } catch {
    return [];
  }
};

/** Компактное представление позиции для контекста модели (экономим токены). */
export const toModelProduct = (product: ProductListItem): Record<string, unknown> => ({
  gost: product.gost ?? null,
  material: product.material ?? null,
  name: product.name,
  package_qty: product.package_qty,
  price_rub: product.client_price || product.base_price,
  size: product.size ?? null,
  sku: product.sku,
  stock_qty: product.stock_qty,
});
