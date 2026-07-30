'use client';

import clsx from 'clsx';

import type { ProductListItem } from '@/core/shared/api/products';

import styles from './CatalogCardsGrid.module.css';
import { CatalogProductCard } from './CatalogProductCard';

type CatalogCardsGridProps = {
  isFavorite: (productId: string) => boolean;
  onAddToCart: (product: ProductListItem) => void;
  onToggleFavorite: (productId: string) => void;
  products: ProductListItem[];
};

export const CatalogCardsGrid = ({
  isFavorite,
  onAddToCart,
  onToggleFavorite,
  products,
}: CatalogCardsGridProps): JSX.Element => {
  return (
    <div className={clsx(styles.grid)}>
      {products.map((product) => (
        <CatalogProductCard
          isFavorite={isFavorite(product.id)}
          key={product.id}
          onAddToCart={() => onAddToCart(product)}
          onToggleFavorite={() => onToggleFavorite(product.id)}
          product={product}
        />
      ))}
    </div>
  );
};
