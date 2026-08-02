'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { Suspense } from 'react';

import { useMediaQuery } from '@/core/shared/lib/useMediaQuery';
import { AppPath } from '@/core/shared/router/paths';
import { Page } from '@/core/shared/ui/Page';
import { toastShown } from '@/core/shared/ui/Toast/model';

import styles from './Catalog.module.css';
import { downloadPriceList } from './lib/exportPrice';
import { useCatalog } from './lib/useCatalog';
import { CatalogB2BTable } from './ui/CatalogB2BTable';
import { BrandFilterChip, OneCUnavailableBanner } from './ui/CatalogBanners';
import { CatalogCardsGrid } from './ui/CatalogCardsGrid';
import { CatalogCategories } from './ui/CatalogCategories';
import { CatalogEmptyState } from './ui/CatalogEmptyState';
import { CatalogFiltersPanel } from './ui/CatalogFilters';
import { CatalogSearchSection } from './ui/CatalogSearchSection';
import { CatalogSkeleton } from './ui/CatalogSkeleton';
import { CatalogToolbar } from './ui/CatalogToolbar';

const is1cUnavailable =
  process.env.NEXT_PUBLIC_CATALOG_1C_UNAVAILABLE === 'true' ||
  process.env.NEXT_PUBLIC_CATALOG_1C_UNAVAILABLE === '1';

/** Совпадает с breakpoint сайдбара каталога (`Catalog.module.css`). */
const CATALOG_COMPACT_QUERY = '(width < 980px)';

const CatalogContent = (): JSX.Element => {
  const showToast = useUnit(toastShown);
  const isCompactLayout = useMediaQuery(CATALOG_COMPACT_QUERY);
  const {
    activeFilterCount,
    categories,
    error,
    favorites,
    filters,
    handleAddToCart,
    isPending,
    patchFilters,
    products,
    resetAll,
    resetFilters,
    setView,
    total,
  } = useCatalog();

  const hasQuery = Boolean(filters.q);
  const showEmpty = !isPending && !error && products.length === 0;
  /** На узких экранах — компактный B2B-список, а не тяжёлые карточки. */
  const effectiveView = isCompactLayout ? 'table' : filters.view;

  const onExportXls = async (): Promise<void> => {
    try {
      await downloadPriceList({
        brandID: filters.brandID,
        categoryID: filters.categoryID,
        gost: filters.gost,
        inStock: filters.inStock,
        material: filters.material,
        q: filters.q,
        size: filters.size,
      });
      showToast({ message: 'Прайс выгружен', tone: 'success' });
    } catch (exportError) {
      showToast({
        message: exportError instanceof Error ? exportError.message : 'Не удалось выгрузить прайс',
        tone: 'error',
      });
    }
  };

  return (
    <div className={clsx(styles.root)}>
      <div className={clsx(styles.container)}>
        <nav aria-label="Хлебные крошки" className={clsx(styles.breadcrumbs)}>
          <Link href={AppPath.Home}>Главная</Link>
          <span>/</span>
          <span>Каталог</span>
        </nav>

        <CatalogToolbar
          canExport={total > 0}
          hideViewToggle={isCompactLayout}
          onExportXls={() => {
            void onExportXls();
          }}
          onPrint={() => window.print()}
          onViewChange={setView}
          total={total}
          view={filters.view}
        />

        {is1cUnavailable ? <OneCUnavailableBanner /> : null}

        {filters.brandID && filters.brandName ? (
          <BrandFilterChip
            brandName={filters.brandName}
            onClear={() => patchFilters({ brandID: undefined, brandName: undefined })}
          />
        ) : null}

        <div className={clsx(styles.layout)}>
          <CatalogCategories
            categories={categories}
            onSelect={(categoryID) => patchFilters({ categoryID })}
            selectedCategoryId={filters.categoryID}
            totalAll={total}
          />

          <div className={clsx(styles.main)}>
            <CatalogFiltersPanel
              activeCount={activeFilterCount}
              filters={filters}
              onChange={patchFilters}
              onReset={resetFilters}
            />

            {isPending ? <CatalogSkeleton view={effectiveView} /> : null}
            {error ? <p className={clsx(styles.error)}>{error}</p> : null}

            {showEmpty ? <CatalogEmptyState onReset={resetAll} /> : null}

            {!isPending && !error && products.length > 0 ? (
              hasQuery ? (
                <CatalogSearchSection
                  isFavorite={favorites.isFavorite}
                  onAddToCart={handleAddToCart}
                  onAddToCartFromCard={(product) =>
                    handleAddToCart(
                      product.id,
                      product.name,
                      product.package_qty,
                      product.package_qty,
                    )
                  }
                  onToggleFavorite={favorites.toggleFavorite}
                  products={products}
                  total={total}
                  view={effectiveView}
                />
              ) : effectiveView === 'table' ? (
                <CatalogB2BTable onAddToCart={handleAddToCart} products={products} />
              ) : (
                <CatalogCardsGrid
                  isFavorite={favorites.isFavorite}
                  onAddToCart={(product) =>
                    handleAddToCart(
                      product.id,
                      product.name,
                      product.package_qty,
                      product.package_qty,
                    )
                  }
                  onToggleFavorite={favorites.toggleFavorite}
                  products={products}
                />
              )
            ) : null}
          </div>
        </div>
      </div>
    </div>
  );
};

export const Catalog = (): JSX.Element => {
  return (
    <Page>
      <Suspense
        fallback={
          <div className={clsx(styles.root)}>
            <div className={clsx(styles.container)}>
              <CatalogSkeleton view="table" />
            </div>
          </div>
        }
      >
        <CatalogContent />
      </Suspense>
    </Page>
  );
};
