'use client';

import { useUnit } from 'effector-react';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useMemo } from 'react';

import { useCart } from '@/core/entities/cart';
import { useFavorites } from '@/core/entities/favorites';
import { normalizeQtyToPackage } from '@/core/shared/lib/formatPrice';
import { getPrimaryProductImageUrl } from '@/core/shared/lib/productImage';
import { toastShown } from '@/core/shared/ui/Toast/model';

import type { CatalogFilters, CatalogViewMode } from './filters';

import {
  $catalogError,
  $catalogFilters,
  $catalogProducts,
  $catalogTotal,
  $categories,
  $isCatalogPending,
  $isCategoriesPending,
  catalogFiltersApplied,
  catalogMounted,
} from '../model';
import {
  catalogFiltersToSearchParams,
  catalogPageCount,
  countActiveFilters,
  parseCatalogFilters,
} from './filters';

export const useCatalog = () => {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { addToCart } = useCart();
  const favorites = useFavorites();

  const [
    filters,
    products,
    total,
    categories,
    isPending,
    isCategoriesPending,
    error,
    mount,
    applyFilters,
    showToast,
  ] = useUnit([
    $catalogFilters,
    $catalogProducts,
    $catalogTotal,
    $categories,
    $isCatalogPending,
    $isCategoriesPending,
    $catalogError,
    catalogMounted,
    catalogFiltersApplied,
    toastShown,
  ]);

  const urlFilters = useMemo(
    () => parseCatalogFilters(new URLSearchParams(searchParams.toString())),
    [searchParams],
  );

  useEffect(() => {
    mount();
  }, [mount]);

  useEffect(() => {
    applyFilters(urlFilters);
  }, [applyFilters, urlFilters]);

  const replaceFilters = (next: CatalogFilters): void => {
    const query = catalogFiltersToSearchParams(next).toString();
    router.replace(query ? `${pathname}?${query}` : pathname, { scroll: false });
  };

  const patchFilters = (patch: Partial<CatalogFilters>): void => {
    const shouldResetPage = patch.page === undefined;
    replaceFilters({
      ...urlFilters,
      ...patch,
      ...(shouldResetPage ? { page: 1 } : {}),
    });
  };

  const resetFilters = (): void => {
    replaceFilters({
      brandID: urlFilters.brandID,
      brandName: urlFilters.brandName,
      categoryID: urlFilters.categoryID,
      page: 1,
      promotionID: urlFilters.promotionID,
      promotionName: urlFilters.promotionName,
      q: urlFilters.q,
      view: urlFilters.view,
    });
  };

  const resetAll = (): void => {
    replaceFilters({ page: 1, view: urlFilters.view });
  };

  const setView = (view: CatalogViewMode): void => {
    replaceFilters({ ...urlFilters, view });
  };

  const setPage = (page: number): void => {
    patchFilters({ page });
  };

  const handleAddToCart = (productID: string, name: string, qty: number, packageQty: number) => {
    if (qty <= 0) {
      showToast({ message: 'Укажите количество', tone: 'error' });

      return;
    }

    const normalizedQty = normalizeQtyToPackage(qty, packageQty);
    const product = products.find((item) => item.id === productID);

    addToCart({
      imageUrl: getPrimaryProductImageUrl(product?.images) ?? undefined,
      name,
      productID,
      qty: normalizedQty,
    });
  };

  return {
    activeFilterCount: countActiveFilters(urlFilters),
    categories,
    error,
    favorites,
    filters: urlFilters,
    handleAddToCart,
    isCategoriesPending,
    isPending,
    pageCount: catalogPageCount(total),
    patchFilters,
    products,
    resetAll,
    resetFilters,
    setPage,
    setView,
    storeFilters: filters,
    total,
  };
};
