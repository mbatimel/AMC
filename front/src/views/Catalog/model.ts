import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Category, ListProductsResult, ProductListItem } from '@/core/shared/api/products';

import { listCategoriesRequest, listProductsRequest } from '@/core/shared/api/products';

import type { CatalogFilters } from './lib/filters';

import { DEFAULT_CATALOG_FILTERS } from './lib/filters';

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

export const $isCatalogPending = createStore(false)
  .on(fetchCatalogProductsFx, () => true)
  .on(fetchCatalogProductsFx.finally, () => false);

export const $catalogError = createStore<null | string>(null)
  .on(fetchCatalogProductsFx, () => null)
  .on(fetchCatalogProductsFx.failData, (_, error) => error.message);

export const $isCategoriesPending = createStore(false)
  .on(fetchCategoriesFx, () => true)
  .on(fetchCategoriesFx.finally, () => false);

sample({
  clock: catalogMounted,
  target: fetchCategoriesFx,
});

sample({
  clock: catalogFiltersApplied,
  target: fetchCatalogProductsFx,
});
