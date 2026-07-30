import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type { Category, ListProductsResult, ProductListItem } from '@/core/shared/api/products';

import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { listCategoriesRequest, listProductsRequest } from '@/core/shared/api/products';

import type { CatalogFilters } from './lib/filters';

import { DEFAULT_CATALOG_FILTERS, toCatalogProductsQueryKey } from './lib/filters';

export const catalogMounted = createEvent();
export const catalogFiltersApplied = createEvent<CatalogFilters>();

export const fetchCatalogProductsFx = createEffect(
  async (filters: CatalogFilters): Promise<ListProductsResult> =>
    listProductsRequest({
      brandID: filters.brandID,
      categoryID: filters.categoryID,
      gost: filters.gost,
      inStock: filters.inStock,
      limit: 100,
      material: filters.material,
      offset: 0,
      q: filters.q,
      size: filters.size,
    }),
);

export const fetchCategoriesFx = createEffect(() => listCategoriesRequest());

export const $catalogFilters = createStore<CatalogFilters>(DEFAULT_CATALOG_FILTERS).on(
  catalogFiltersApplied,
  (_, filters) => filters,
);

export const $catalogProducts = createStore<ProductListItem[]>([]).on(
  fetchCatalogProductsFx.doneData,
  (_, result) => result.items,
);

export const $catalogTotal = createStore(0).on(
  fetchCatalogProductsFx.doneData,
  (_, result) => result.pagination.total,
);

export const $categories = createStore<Category[]>([]).on(
  fetchCategoriesFx.doneData,
  (_, categories) => categories,
);

export const $isCatalogPending = fetchCatalogProductsFx.pending;

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

/** Последний успешно запрошенный ключ продуктов (API-поля). */
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
