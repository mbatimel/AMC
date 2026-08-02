import type { ProductListItem } from '@/core/shared/api/products';

export type CatalogFilters = {
  brandID?: string;
  brandName?: string;
  categoryID?: string;
  gost?: string;
  inStock?: boolean;
  material?: string;
  promotionID?: string;
  promotionName?: string;
  q?: string;
  size?: string;
  view: CatalogViewMode;
};

export type CatalogViewMode = 'cards' | 'table';

export const DEFAULT_CATALOG_FILTERS: CatalogFilters = {
  view: 'table',
};

export const parseCatalogFilters = (params: URLSearchParams): CatalogFilters => {
  const view = params.get('view');

  return {
    brandID: params.get('brandID') || undefined,
    brandName: params.get('brandName') || params.get('brand') || undefined,
    categoryID: params.get('categoryID') || undefined,
    gost: params.get('gost') || undefined,
    inStock: params.get('inStock') === 'true' ? true : undefined,
    material: params.get('material') || undefined,
    promotionID: params.get('promotionID') || undefined,
    promotionName: params.get('promotionName') || undefined,
    q: params.get('q') || undefined,
    size: params.get('size') || undefined,
    view: view === 'cards' ? 'cards' : 'table',
  };
};

export const catalogFiltersToSearchParams = (filters: CatalogFilters): URLSearchParams => {
  const params = new URLSearchParams();

  if (filters.q) {
    params.set('q', filters.q);
  }

  if (filters.categoryID) {
    params.set('categoryID', filters.categoryID);
  }

  if (filters.brandID) {
    params.set('brandID', filters.brandID);
  }

  if (filters.brandName) {
    params.set('brandName', filters.brandName);
  }

  if (filters.promotionID) {
    params.set('promotionID', filters.promotionID);
  }

  if (filters.promotionName) {
    params.set('promotionName', filters.promotionName);
  }

  if (filters.material) {
    params.set('material', filters.material);
  }

  if (filters.size) {
    params.set('size', filters.size);
  }

  if (filters.gost) {
    params.set('gost', filters.gost);
  }

  if (filters.inStock) {
    params.set('inStock', 'true');
  }

  if (filters.view === 'cards') {
    params.set('view', 'cards');
  }

  return params;
};

export const countActiveFilters = (filters: CatalogFilters): number => {
  let count = 0;

  if (filters.material) {
    count += 1;
  }

  if (filters.size) {
    count += 1;
  }

  if (filters.gost) {
    count += 1;
  }

  if (filters.inStock) {
    count += 1;
  }

  return count;
};

/** Ключ загрузки: для акции — только promotionID; иначе поля listProducts. */
export const toCatalogProductsQueryKey = (filters: CatalogFilters): string => {
  if (filters.promotionID) {
    return `promo\u001f${filters.promotionID}`;
  }

  return [
    filters.brandID ?? '',
    filters.categoryID ?? '',
    filters.gost ?? '',
    filters.inStock ? '1' : '0',
    filters.material ?? '',
    filters.q ?? '',
    filters.size ?? '',
  ].join('\u001f');
};

const matchesQuery = (product: ProductListItem, q: string): boolean => {
  const needle = q.trim().toLowerCase();

  if (!needle) {
    return true;
  }

  return [product.name, product.sku, product.gost, product.brand_name, product.category_name]
    .filter(Boolean)
    .some((value) => value!.toLowerCase().includes(needle));
};

/** Клиентские фильтры поверх полного набора товаров акции. */
export const applyClientCatalogFilters = (
  products: ProductListItem[],
  filters: CatalogFilters,
): ProductListItem[] =>
  products.filter((product) => {
    if (filters.brandID && product.brand_id !== filters.brandID) {
      return false;
    }

    if (filters.categoryID && product.category_id !== filters.categoryID) {
      return false;
    }

    if (filters.material && product.material !== filters.material) {
      return false;
    }

    if (filters.size && product.size !== filters.size) {
      return false;
    }

    if (filters.gost && product.gost !== filters.gost) {
      return false;
    }

    if (filters.inStock && product.stock_qty <= 0) {
      return false;
    }

    if (filters.q && !matchesQuery(product, filters.q)) {
      return false;
    }

    return true;
  });
