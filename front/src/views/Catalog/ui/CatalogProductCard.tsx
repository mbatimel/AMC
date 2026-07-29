'use client';

import { Button, Chip } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useState } from 'react';

import type { ProductListItem } from '@/core/shared/api/products';

import { formatPrice } from '@/core/shared/lib/formatPrice';
import { getStockLevel } from '@/core/shared/lib/stock';
import { getProductPath } from '@/core/shared/router/paths';
import { ProductImageFallback } from '@/core/shared/ui/ProductImageFallback';

import styles from './CatalogProductCard.module.css';

type CatalogProductCardProps = {
  isFavorite: boolean;
  onAddToCart: () => void;
  onToggleFavorite: () => void;
  product: ProductListItem;
};

export const CatalogProductCard = ({
  isFavorite,
  onAddToCart,
  onToggleFavorite,
  product,
}: CatalogProductCardProps): JSX.Element => {
  const image = product.images?.find((item) => item.is_primary) ?? product.images?.[0];
  const [failed, setFailed] = useState(false);
  const stockLevel = getStockLevel(product.stock_qty);
  const isOut = stockLevel === 'out';
  const hasDiscount = product.discount_percent > 0;

  return (
    <article className={clsx(styles.root)}>
      <div className={clsx(styles.media)}>
        <Chip
          className={clsx(
            styles.stock,
            stockLevel === 'out' && styles.stockOut,
            stockLevel === 'low' && styles.stockLow,
          )}
          color={isOut ? 'danger' : stockLevel === 'low' ? 'warning' : 'success'}
          size="sm"
        >
          <Chip.Label>{isOut ? 'Нет в наличии' : `В наличии: ${product.stock_qty}`}</Chip.Label>
        </Chip>
        <Button
          aria-label={isFavorite ? 'Убрать из избранного' : 'В избранное'}
          className={clsx(styles.favorite, isFavorite && styles.favoriteActive)}
          isIconOnly
          onPress={onToggleFavorite}
          size="sm"
          variant="outline"
        >
          ★
        </Button>
        <div className={clsx(styles.imageWrap)}>
          {!image || failed ? (
            <ProductImageFallback categoryName={product.category_name} />
          ) : (
            <img
              alt={product.name}
              className={clsx(styles.image)}
              onError={() => setFailed(true)}
              src={image.url}
            />
          )}
        </div>
        {hasDiscount ? (
          <Chip className={clsx(styles.discount)} color="warning" size="sm">
            <Chip.Label>-{Math.round(product.discount_percent)}%</Chip.Label>
          </Chip>
        ) : null}
      </div>

      <div className={clsx(styles.body)}>
        <p className={clsx(styles.meta)}>
          {product.sku}
          {product.gost ? ` • ${product.gost}` : ''}
        </p>
        <h3 className={clsx(styles.title)}>{product.name}</h3>
        <p className={clsx(styles.specs)}>
          {[product.material, product.size, `уп. ${product.package_qty} шт`]
            .filter(Boolean)
            .join(' • ')}
        </p>

        <div className={clsx(styles.priceBlock)}>
          <span className={clsx(styles.price, hasDiscount && styles.pricePromo)}>
            {formatPrice(product.client_price)}
          </span>
          {hasDiscount ? (
            <span className={clsx(styles.oldPrice)}>{formatPrice(product.base_price)}</span>
          ) : null}
          <span className={clsx(styles.priceHint)}>Ваша оптовая цена</span>
        </div>

        <div className={clsx(styles.actions)}>
          <Link className={clsx(styles.details)} href={getProductPath(product.id)}>
            Подробнее
          </Link>
          <Button
            className={clsx(styles.cartButton)}
            isDisabled={isOut}
            onPress={onAddToCart}
            size="sm"
            variant="primary"
          >
            {isOut ? 'Нет' : 'В корзину'}
          </Button>
        </div>
      </div>
    </article>
  );
};
