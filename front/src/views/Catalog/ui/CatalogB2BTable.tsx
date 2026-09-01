'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useState } from 'react';

import type { ProductListItem } from '@/core/shared/api/products';

import { IconCart, IconFavorite, IconPriceTag } from '@/core/shared/icons';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { getPrimaryProductImage } from '@/core/shared/lib/productImage';
import { getStockLevel } from '@/core/shared/lib/stock';
import { getProductPath } from '@/core/shared/router/paths';
import { ProductImageFallback } from '@/core/shared/ui/ProductImageFallback';
import { QuantityStepper } from '@/core/shared/ui/QuantityStepper';

import styles from './CatalogB2BTable.module.css';

type CatalogB2BTableProps = {
  isFavorite: (productId: string) => boolean;
  isPreviouslyOrdered: (productId: string) => boolean;
  onAddToCart: (productID: string, name: string, qty: number, packageQty: number) => void;
  onToggleFavorite: (productId: string) => void;
  products: ProductListItem[];
};

const EMPTY = '—';

const formatPlainNumber = (value: number): string =>
  new Intl.NumberFormat('ru-RU', {
    maximumFractionDigits: 2,
    minimumFractionDigits: 2,
  }).format(value);

const displayValue = (value: null | string | undefined): string => {
  const trimmed = value?.trim();

  return trimmed ? trimmed : EMPTY;
};

const formatCompactMeta = (product: ProductListItem): string => {
  const parts = [product.gost, product.material, product.size]
    .map((value) => value?.trim())
    .filter(Boolean) as string[];

  parts.push(`уп. ${product.package_qty} шт`);

  return parts.join(' · ');
};

const ProductThumb = ({ product }: { product: ProductListItem }): JSX.Element => {
  const image = getPrimaryProductImage(product.images);
  const [failed, setFailed] = useState(false);

  if (!image || failed) {
    return (
      <ProductImageFallback
        categoryName={product.category_name}
        className={clsx(styles.thumbFallback)}
        label=""
      />
    );
  }

  return (
    <img
      alt={product.name}
      className={clsx(styles.thumbImage)}
      onError={() => setFailed(true)}
      src={image.url}
    />
  );
};

export const CatalogB2BTable = ({
  isFavorite,
  isPreviouslyOrdered,
  onAddToCart,
  onToggleFavorite,
  products,
}: CatalogB2BTableProps): JSX.Element => {
  const [qtyById, setQtyById] = useState<Record<string, number>>({});

  return (
    <div className={clsx(styles.root)}>
      <div className={clsx(styles.grid)}>
        <div aria-hidden className={clsx(styles.head)}>
          <span className={clsx(styles.headProduct)}>Товар</span>
          <span className={clsx(styles.headSpec)}>ГОСТ</span>
          <span className={clsx(styles.headSpec)}>Материал</span>
          <span className={clsx(styles.headSpec)}>Размер</span>
          <span className={clsx(styles.headSpec)}>Уп.</span>
          <span className={clsx(styles.headSpec)}>Остаток</span>
          <span className={clsx(styles.headPrice)}>Цена</span>
          <span className={clsx(styles.headActions)}>Кол-во</span>
        </div>

        <ul className={clsx(styles.list)}>
          {products.map((product) => {
            const stockLevel = getStockLevel(product.stock_qty);
            const isOut = stockLevel === 'out';
            const qty = qtyById[product.id] ?? product.package_qty;
            const showDiscount =
              product.discount_percent > 0 ||
              (product.base_price > 0 && product.base_price > product.client_price);

            return (
              <li className={clsx(styles.row)} key={product.id}>
                <div className={clsx(styles.product)}>
                  <div className={clsx(styles.thumb)}>
                    <ProductThumb product={product} />
                  </div>
                  <div className={clsx(styles.info)}>
                    {product.sku ? <p className={clsx(styles.sku)}>{product.sku}</p> : null}
                    <Link className={clsx(styles.nameLink)} href={getProductPath(product.id)}>
                      {product.name}
                    </Link>
                    {isFavorite(product.id) || isPreviouslyOrdered(product.id) ? (
                      <p className={clsx(styles.marks)}>
                        {isFavorite(product.id) ? (
                          <span className={clsx(styles.mark)}>В моём избранном</span>
                        ) : null}
                        {isPreviouslyOrdered(product.id) ? (
                          <span className={clsx(styles.mark)}>Ранее заказывали</span>
                        ) : null}
                      </p>
                    ) : null}
                    <p className={clsx(styles.compactMeta)}>{formatCompactMeta(product)}</p>
                    <span
                      className={clsx(
                        styles.stockBadge,
                        stockLevel === 'out' && styles.stockOut,
                        stockLevel === 'low' && styles.stockLow,
                      )}
                    >
                      {isOut ? 'Нет в наличии' : `В наличии: ${product.stock_qty}`}
                    </span>
                  </div>
                </div>

                <span className={clsx(styles.spec, styles.gost)}>{displayValue(product.gost)}</span>
                <span className={clsx(styles.spec, styles.material)}>
                  {displayValue(product.material)}
                </span>
                <span className={clsx(styles.spec, styles.size)}>{displayValue(product.size)}</span>
                <span className={clsx(styles.spec, styles.pack)}>{product.package_qty} шт</span>
                <span
                  className={clsx(
                    styles.spec,
                    styles.stock,
                    stockLevel === 'out' && styles.stockOut,
                    stockLevel === 'low' && styles.stockLow,
                  )}
                >
                  {isOut ? '0' : product.stock_qty}
                </span>

                <div className={clsx(styles.priceBlock)}>
                  {showDiscount ? (
                    <>
                      <span className={clsx(styles.basePrice)}>
                        {formatPlainNumber(product.base_price)} ₽
                      </span>
                      <span className={clsx(styles.priceBadge)}>
                        <IconPriceTag currentColor="currentColor" height={12} width={12} />
                        {formatPrice(product.client_price)}
                      </span>
                    </>
                  ) : (
                    <span className={clsx(styles.price)}>{formatPrice(product.client_price)}</span>
                  )}
                </div>

                <div className={clsx(styles.actions)}>
                  <button
                    aria-label={
                      isFavorite(product.id)
                        ? 'Убрать из моего избранного'
                        : 'Добавить в моё избранное'
                    }
                    className={clsx(
                      styles.favoriteButton,
                      isFavorite(product.id) && styles.favoriteButtonActive,
                    )}
                    onClick={() => onToggleFavorite(product.id)}
                    title={
                      isFavorite(product.id)
                        ? 'Убрать из моего избранного'
                        : 'Добавить в моё избранное'
                    }
                    type="button"
                  >
                    <IconFavorite currentColor="currentColor" height={16} width={16} />
                  </button>
                  <QuantityStepper
                    className={clsx(styles.stepper)}
                    disabled={isOut}
                    onChange={(value) =>
                      setQtyById((current) => ({ ...current, [product.id]: value }))
                    }
                    step={product.package_qty}
                    value={qty}
                  />
                  <Button
                    className={clsx(styles.cartButton)}
                    isDisabled={isOut}
                    onPress={() => onAddToCart(product.id, product.name, qty, product.package_qty)}
                    size="sm"
                  >
                    {isOut ? (
                      'Нет'
                    ) : (
                      <>
                        <IconCart currentColor="currentColor" height={14} width={14} />
                        <span className={clsx(styles.cartButtonLabel)}>В корзину</span>
                      </>
                    )}
                  </Button>
                </div>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
};
