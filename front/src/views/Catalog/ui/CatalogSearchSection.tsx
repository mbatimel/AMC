'use client';

import clsx from 'clsx';

import type { ProductListItem } from '@/core/shared/api/products';

import { CatalogB2BTable } from './CatalogB2BTable';
import { CatalogCardsGrid } from './CatalogCardsGrid';
import styles from './CatalogSearchSection.module.css';

type CatalogSearchSectionProps = {
  isFavorite: (productId: string) => boolean;
  onAddToCart: (productID: string, name: string, qty: number, packageQty: number) => void;
  onAddToCartFromCard: (product: ProductListItem) => void;
  onToggleFavorite: (productId: string) => void;
  products: ProductListItem[];
  total: number;
  view: 'cards' | 'table';
};

export const CatalogSearchSection = ({
  isFavorite,
  onAddToCart,
  onAddToCartFromCard,
  onToggleFavorite,
  products,
  total,
  view,
}: CatalogSearchSectionProps): JSX.Element => {
  const [bestMatch, ...rest] = products;

  if (!bestMatch) {
    return <></>;
  }

  return (
    <div className={clsx(styles.root)}>
      <section className={clsx(styles.section)}>
        <h2 className={clsx(styles.title)}>Лучшее совпадение</h2>
        {view === 'table' ? (
          <CatalogB2BTable onAddToCart={onAddToCart} products={[bestMatch]} />
        ) : (
          <CatalogCardsGrid
            isFavorite={isFavorite}
            onAddToCart={onAddToCartFromCard}
            onToggleFavorite={onToggleFavorite}
            products={[bestMatch]}
          />
        )}
      </section>

      {rest.length > 0 ? (
        <section className={clsx(styles.section)}>
          <h2 className={clsx(styles.title)}>
            Похожие позиции и категория — ещё {Math.max(total - 1, rest.length)}
          </h2>
          {view === 'table' ? (
            <CatalogB2BTable onAddToCart={onAddToCart} products={rest} />
          ) : (
            <CatalogCardsGrid
              isFavorite={isFavorite}
              onAddToCart={onAddToCartFromCard}
              onToggleFavorite={onToggleFavorite}
              products={rest}
            />
          )}
        </section>
      ) : null}
    </div>
  );
};
