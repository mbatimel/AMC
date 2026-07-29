'use client';

import { useUnit } from 'effector-react';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useMemo } from 'react';

import { useCart } from '@/core/entities/cart';
import { useFavorites } from '@/core/entities/favorites';
import { normalizeQtyToPackage } from '@/core/shared/lib/formatPrice';
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
import { catalogFiltersToSearchParams, countActiveFilters, parseCatalogFilters } from './filters';

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
    replaceFilters({ ...urlFilters, ...patch });
  };

  const resetFilters = (): void => {
    replaceFilters({
      brandID: urlFilters.brandID,
      brandName: urlFilters.brandName,
      categoryID: urlFilters.categoryID,
      q: urlFilters.q,
      view: urlFilters.view,
    });
  };

  const resetAll = (): void => {
    replaceFilters({ view: urlFilters.view });
  };

  const setView = (view: CatalogViewMode): void => {
    patchFilters({ view });
  };

  const handleAddToCart = (productID: string, name: string, qty: number, packageQty: number) => {
    if (qty <= 0) {
      showToast({ message: 'Укажите количество', tone: 'error' });

      return;
    }

    const normalizedQty = normalizeQtyToPackage(qty, packageQty);

    addToCart({ name, productID, qty: normalizedQty });
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
    patchFilters,
    products,
    resetAll,
    resetFilters,
    setView,
    storeFilters: filters,
    total,
  };
};
