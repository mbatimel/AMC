import { assertApiSuccess, omitEmptyParams, parseApiErrorMessage } from './parseApiError';

export type Brand = {
  id: string;
  name: string;
  slug?: string;
};

export type Category = {
  id: string;
  name: string;
  parent_id?: string;
  slug?: string;
};

export type ListProductsParams = {
  brandID?: string;
  categoryID?: string;
  gost?: string;
  inStock?: boolean;
  limit?: number;
  material?: string;
  offset?: number;
  q?: string;
  size?: string;
  sort?: string;
};

export type ListProductsResult = {
  items: ProductListItem[];
  pagination: Pagination;
};

export type Pagination = {
  limit: number;
  offset: number;
  total: number;
};

export type Product = ProductListItem & {
  description?: string;
};

export type ProductImage = {
  alt?: string;
  id: string;
  is_primary?: boolean;
  sort_order?: number;
  url: string;
};

export type ProductListItem = {
  base_price: number;
  brand_id?: string;
  brand_name?: string;
  category_id?: string;
  category_name?: string;
  client_price: number;
  discount_percent: number;
  gost?: string;
  id: string;
  images?: null | ProductImage[];
  is_published?: boolean;
  material?: string;
  name: string;
  package_qty: number;
  size?: string;
  sku: string;
  stock_qty: number;
};

export class ProductsApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ProductsApiError';
    this.status = status;
  }
}

const asNumber = (value: unknown, fallback = 0): number =>
  typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const asString = (value: unknown, fallback = ''): string =>
  typeof value === 'string' ? value : fallback;

const parseImages = (value: unknown): ProductImage[] => {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.flatMap((item) => {
    if (typeof item !== 'object' || item === null) {
      return [];
    }

    const record = item as Record<string, unknown>;
    const id = asString(record.id);
    const url = asString(record.url);

    if (!id || !url) {
      return [];
    }

    return [
      {
        alt: asString(record.alt) || undefined,
        id,
        is_primary: record.is_primary === true,
        sort_order: asNumber(record.sort_order),
        url,
      },
    ];
  });
};

const parseProductListItem = (value: unknown): null | ProductListItem => {
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
    images: parseImages(record.images),
    is_published: record.is_published !== false,
    material: asString(record.material) || undefined,
    name,
    package_qty: Math.max(1, asNumber(record.package_qty, 1)),
    size: asString(record.size) || undefined,
    sku,
    stock_qty: asNumber(record.stock_qty),
  };
};

const parsePagination = (value: unknown, itemsLength: number): Pagination => {
  if (typeof value === 'object' && value !== null) {
    const record = value as Record<string, unknown>;

    return {
      limit: asNumber(record.limit, itemsLength),
      offset: asNumber(record.offset),
      total: asNumber(record.total, itemsLength),
    };
  }

  return { limit: itemsLength, offset: 0, total: itemsLength };
};

const parseListProducts = (data: unknown): ListProductsResult => {
  const record = assertApiSuccess(data, 'Некорректный ответ сервера');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new ProductsApiError(500, 'Некорректный ответ сервера');
  }

  const payloadRecord = payload as Record<string, unknown>;
  const rawItems = Array.isArray(payloadRecord.items) ? payloadRecord.items : [];
  const items = rawItems.flatMap((item) => {
    const parsed = parseProductListItem(item);

    return parsed ? [parsed] : [];
  });

  return {
    items,
    pagination: parsePagination(payloadRecord.pagination, items.length),
  };
};

const parseProduct = (data: unknown): Product => {
  const record = assertApiSuccess(data, 'Не удалось загрузить товар');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new ProductsApiError(500, 'Некорректный ответ сервера');
  }

  const payloadRecord = payload as Record<string, unknown>;
  const productPayload = payloadRecord.product ?? payload;
  const product = parseProductListItem(productPayload);

  if (!product) {
    throw new ProductsApiError(500, 'Некорректный ответ сервера');
  }

  const source =
    typeof productPayload === 'object' && productPayload !== null
      ? (productPayload as Record<string, unknown>)
      : {};

  return {
    ...product,
    description: asString(source.description) || undefined,
  };
};

const parseNamedList = <T extends { id: string; name: string }>(
  data: unknown,
  mapItem: (record: Record<string, unknown>) => null | T,
  fallback: string,
): T[] => {
  const record = assertApiSuccess(data, fallback);
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new ProductsApiError(500, fallback);
  }

  const payloadRecord = payload as Record<string, unknown>;
  const rawItems = Array.isArray(payloadRecord.items)
    ? payloadRecord.items
    : Array.isArray(payload)
      ? payload
      : [];

  return rawItems.flatMap((item) => {
    if (typeof item !== 'object' || item === null) {
      return [];
    }

    const mapped = mapItem(item as Record<string, unknown>);

    return mapped ? [mapped] : [];
  });
};

export const listProductsRequest = async (
  params: ListProductsParams = {},
): Promise<ListProductsResult> => {
  const query = omitEmptyParams({
    brandID: params.brandID,
    categoryID: params.categoryID,
    gost: params.gost,
    inStock: params.inStock,
    limit: params.limit ?? 50,
    material: params.material,
    offset: params.offset ?? 0,
    q: params.q,
    size: params.size,
    sort: params.sort,
  });

  const response = await fetch(`/api/v1/products${query}`);

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить каталог'),
    );
  }

  const data: unknown = await response.json();

  return parseListProducts(data);
};

export const fetchAllProductsRequest = async (
  params: Omit<ListProductsParams, 'limit' | 'offset'> = {},
): Promise<ProductListItem[]> => {
  const pageSize = 100;
  let offset = 0;
  let total = Infinity;
  const items: ProductListItem[] = [];

  while (offset < total) {
    const page = await listProductsRequest({ ...params, limit: pageSize, offset });

    items.push(...page.items);
    total = page.pagination.total;
    offset += pageSize;

    if (page.items.length === 0) {
      break;
    }
  }

  return items;
};

export const getProductRequest = async (productId: string): Promise<Product> => {
  const response = await fetch(`/api/v1/products/${productId}`);

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить товар'),
    );
  }

  const data: unknown = await response.json();

  return parseProduct(data);
};

export const listCategoriesRequest = async (): Promise<Category[]> => {
  const response = await fetch('/api/v1/categories?limit=100&offset=0');

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить категории'),
    );
  }

  const data: unknown = await response.json();

  return parseNamedList(
    data,
    (record) => {
      const id = asString(record.id);
      const name = asString(record.name);

      if (!id || !name) {
        return null;
      }

      return {
        id,
        name,
        parent_id: asString(record.parent_id) || undefined,
        slug: asString(record.slug) || undefined,
      };
    },
    'Не удалось загрузить категории',
  );
};

export const listBrandsRequest = async (): Promise<Brand[]> => {
  const response = await fetch('/api/v1/brands?limit=100&offset=0');

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить бренды'),
    );
  }

  const data: unknown = await response.json();

  return parseNamedList(
    data,
    (record) => {
      const id = asString(record.id);
      const name = asString(record.name);

      if (!id || !name) {
        return null;
      }

      return {
        id,
        name,
        slug: asString(record.slug) || undefined,
      };
    },
    'Не удалось загрузить бренды',
  );
};

/* ------------------------------------------------------------------ *
 * Админские операции над каталогом (ProductsAPI, требуют X-User-Id)
 * ------------------------------------------------------------------ */

export type ProductWritePayload = {
  basePrice: number;
  brandID?: string;
  categoryID?: string;
  description: string;
  discountPercent: number;
  gost: string;
  images: { alt?: string; is_primary?: boolean; sort_order?: number; url: string }[];
  isPublished: boolean;
  material: string;
  name: string;
  packageQty: number;
  size: string;
  sku: string;
  stockQty: number;
};

const productWriteBody = (
  payload: ProductWritePayload,
  includeImages = true,
): Record<string, unknown> => ({
  basePrice: payload.basePrice,
  brandID: payload.brandID || undefined,
  categoryID: payload.categoryID || undefined,
  description: payload.description,
  discountPercent: payload.discountPercent,
  gost: payload.gost,
  images: includeImages ? payload.images : undefined,
  isPublished: payload.isPublished,
  material: payload.material,
  name: payload.name,
  packageQty: payload.packageQty,
  size: payload.size,
  sku: payload.sku,
  stockQty: payload.stockQty,
});

const adminHeaders = (userId: string): HeadersInit => ({
  'Content-Type': 'application/json',
  'X-User-Id': userId,
});

export const createProductRequest = async (
  userId: string,
  payload: ProductWritePayload,
): Promise<void> => {
  const response = await fetch('/api/v1/products', {
    body: JSON.stringify(productWriteBody(payload)),
    headers: adminHeaders(userId),
    method: 'POST',
  });

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось создать товар'),
    );
  }

  assertApiSuccess(await response.json(), 'Не удалось создать товар');
};

export const updateProductRequest = async (
  userId: string,
  productId: string,
  payload: ProductWritePayload,
): Promise<void> => {
  const response = await fetch(`/api/v1/products/${productId}`, {
    body: JSON.stringify(productWriteBody(payload, false)),
    headers: adminHeaders(userId),
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось сохранить товар'),
    );
  }

  assertApiSuccess(await response.json(), 'Не удалось сохранить товар');
};

export type ProductImageBatchResult = {
  errorText?: string;
  fileName: string;
  image?: ProductImage;
  sku: string;
  success: boolean;
};

export const uploadProductImageRequest = async (
  userId: string,
  productId: string,
  file: File,
): Promise<ProductImage> => {
  const body = new FormData();

  body.append('file', file);
  const response = await fetch(`/api/v1/products/${productId}/images`, {
    body,
    headers: { 'X-User-Id': userId },
    method: 'POST',
  });

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить изображение'),
    );
  }

  const record = assertApiSuccess(await response.json(), 'Не удалось загрузить изображение');
  const payload = record.data as { image?: unknown };
  const image = parseImages([payload.image])[0];

  if (!image) {
    throw new ProductsApiError(500, 'Некорректный ответ сервера');
  }

  return image;
};

export const deleteProductImageRequest = async (
  userId: string,
  productId: string,
  imageId: string,
): Promise<void> => {
  const response = await fetch(`/api/v1/products/${productId}/images/${imageId}`, {
    headers: { 'X-User-Id': userId },
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось удалить изображение'),
    );
  }

  assertApiSuccess(await response.json(), 'Не удалось удалить изображение');
};

export const uploadProductImagesBatchRequest = async (
  userId: string,
  files: File[],
): Promise<ProductImageBatchResult[]> => {
  const body = new FormData();

  files.forEach((file) => body.append('files', file));
  const response = await fetch('/api/v1/products/images/batch', {
    body,
    headers: { 'X-User-Id': userId },
    method: 'POST',
  });

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось загрузить фотографии'),
    );
  }

  const record = assertApiSuccess(await response.json(), 'Не удалось загрузить фотографии');
  const payload = record.data as { items?: ProductImageBatchResult[] };

  return Array.isArray(payload.items) ? payload.items : [];
};

export const deleteProductRequest = async (userId: string, productId: string): Promise<void> => {
  const response = await fetch(`/api/v1/products/${productId}`, {
    headers: { 'X-User-Id': userId },
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new ProductsApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось удалить товар'),
    );
  }

  assertApiSuccess(await response.json(), 'Не удалось удалить товар');
};
