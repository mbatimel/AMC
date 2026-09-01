import { assertApiSuccess, parseApiErrorMessage } from './parseApiError';

export type FavoriteItem = {
  product_id: string;
};

const withUserHeaders = (userId: string, json = false): HeadersInit => ({
  ...(json ? { 'Content-Type': 'application/json' } : {}),
  'X-User-Id': userId,
});

const parseProductIds = (data: unknown): string[] => {
  const record = assertApiSuccess(data, 'Не удалось загрузить избранное');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    return [];
  }

  const items = (payload as { items?: unknown }).items;

  if (!Array.isArray(items)) {
    return [];
  }

  return items.flatMap((item) => {
    if (typeof item !== 'object' || item === null) {
      return [];
    }

    const productId = (item as { product_id?: unknown }).product_id;

    return typeof productId === 'string' && productId.length > 0 ? [productId] : [];
  });
};

export const listFavoritesRequest = async (userId: string): Promise<string[]> => {
  const response = await fetch('/api/v1/users/favorites', {
    headers: withUserHeaders(userId),
  });

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, 'Не удалось загрузить избранное'));
  }

  return parseProductIds(await response.json());
};

export const addFavoriteRequest = async (userId: string, productID: string): Promise<void> => {
  const response = await fetch('/api/v1/users/favorites', {
    body: JSON.stringify({ productID }),
    headers: withUserHeaders(userId, true),
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, 'Не удалось добавить в избранное'));
  }
};

export const deleteFavoriteRequest = async (userId: string, productID: string): Promise<void> => {
  const response = await fetch('/api/v1/users/favorites', {
    body: JSON.stringify({ productIDs: [productID] }),
    headers: withUserHeaders(userId, true),
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, 'Не удалось убрать из избранного'));
  }
};
