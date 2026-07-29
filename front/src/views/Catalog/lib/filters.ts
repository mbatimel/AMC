export type CatalogFilters = {
  brandID?: string;
  brandName?: string;
  categoryID?: string;
  gost?: string;
  inStock?: boolean;
  material?: string;
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
    brandName: params.get('brandName') || undefined,
    categoryID: params.get('categoryID') || undefined,
    gost: params.get('gost') || undefined,
    inStock: params.get('inStock') === 'true' ? true : undefined,
    material: params.get('material') || undefined,
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
