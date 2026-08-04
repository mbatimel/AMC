import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type { Category, ListProductsResult, ProductListItem } from '@/core/shared/api/products';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import {
  getProductRequest,
  listCategoriesRequest,
  listProductsRequest,
} from '@/core/shared/api/products';
import { getPromotionRequest } from '@/core/shared/api/promotions';

import type { CatalogFilters } from './lib/filters';

import {
  applyClientCatalogFilters,
  CATALOG_PAGE_SIZE,
  DEFAULT_CATALOG_FILTERS,
  paginateProducts,
  toCatalogProductsQueryKey,
} from './lib/filters';

export const catalogMounted = createEvent();
export const catalogFiltersApplied = createEvent<CatalogFilters>();

const fetchPromotionProducts = async (promotionID: string): Promise<ListProductsResult> => {
  const promotion = await getPromotionRequest(promotionID);
  const productIds = promotion.products.map((item) => item.product_id);

  const settled = await Promise.allSettled(productIds.map((id) => getProductRequest(id)));
  const items = settled.flatMap((result) =>
    result.status === 'fulfilled' ? [result.value as ProductListItem] : [],
  );

  return {
    items,
    pagination: { limit: items.length, offset: 0, total: items.length },
  };
};

export const fetchCatalogProductsFx = createEffect(
  async (filters: CatalogFilters): Promise<ListProductsResult> => {
    if (filters.promotionID) {
      return fetchPromotionProducts(filters.promotionID);
    }

    const page = Math.max(1, filters.page);
    const offset = (page - 1) * CATALOG_PAGE_SIZE;

    return listProductsRequest({
      brandID: filters.brandID,
      categoryID: filters.categoryID,
      gost: filters.gost,
      inStock: filters.inStock,
      limit: CATALOG_PAGE_SIZE,
      material: filters.material,
      offset,
      q: filters.q,
      size: filters.size,
    });
  },
);

export const fetchCategoriesFx = createEffect(() => listCategoriesRequest());

export const $catalogFilters = createStore<CatalogFilters>(DEFAULT_CATALOG_FILTERS).on(
  catalogFiltersApplied,
  (_, filters) => filters,
);

const $catalogRawProducts = createStore<ProductListItem[]>([]).on(
  fetchCatalogProductsFx.doneData,
  (_, result) => result.items,
);

const $catalogRawTotal = createStore(0).on(
  fetchCatalogProductsFx.doneData,
  (_, result) => result.pagination.total,
);

export const $catalogProducts = combine(
  $catalogRawProducts,
  $catalogFilters,
  (products, filters) => {
    if (!filters.promotionID) {
      return products;
    }

    return paginateProducts(applyClientCatalogFilters(products, filters), filters.page);
  },
);

export const $catalogTotal = combine(
  $catalogRawProducts,
  $catalogRawTotal,
  $catalogFilters,
  (products, rawTotal, filters) =>
    filters.promotionID ? applyClientCatalogFilters(products, filters).length : rawTotal,
);

export const $categories = createStore<Category[]>([]).on(
  fetchCategoriesFx.doneData,
  (_, categories) => categories,
);

/** До первого запроса — true, чтобы не мелькал empty state до useEffect. */
export const $isCatalogPending = createStore(true).on(
  fetchCatalogProductsFx.pending,
  (_, pending) => pending,
);

export const $catalogError = createStore<null | string>(null)
  .on(fetchCatalogProductsFx, () => null)
  .on(fetchCatalogProductsFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить каталог'),
  );

export const $isCategoriesPending = fetchCategoriesFx.pending;

/** Categories: один GET на сессию вкладки (пока store жив). */
const $isCategoriesFetched = createStore(false)
  .on(fetchCategoriesFx, () => true)
  .on(fetchCategoriesFx.fail, () => false);

/** Последний успешно запрошенный ключ продуктов (API-поля / promo id). */
const $productsQueryKey = createStore<null | string>(null)
  .on(fetchCatalogProductsFx, (_, filters) => toCatalogProductsQueryKey(filters))
  .on(fetchCatalogProductsFx.fail, () => null);

const $canFetchCategories = combine(
  $isCategoriesFetched,
  fetchCategoriesFx.pending,
  (fetched, pending) => !fetched && !pending,
);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: catalogMounted,
  source: $canFetchCategories,
  filter: Boolean,
  target: fetchCategoriesFx,
});

sample({
  clock: catalogFiltersApplied,
  source: {
    key: $productsQueryKey,
    pending: fetchCatalogProductsFx.pending,
  },
  filter: ({ key, pending }, filters) => !pending && toCatalogProductsQueryKey(filters) !== key,
  fn: (_, filters) => filters,
  target: fetchCatalogProductsFx,
});

/* eslint-enable perfectionist/sort-objects */
