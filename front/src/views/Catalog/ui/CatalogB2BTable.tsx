'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useState } from 'react';

import type { ProductListItem } from '@/core/shared/api/products';

import { IconPlus } from '@/core/shared/icons';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { getStockLevel } from '@/core/shared/lib/stock';
import { getProductPath } from '@/core/shared/router/paths';
import { ProductImageFallback } from '@/core/shared/ui/ProductImageFallback';
import { QuantityStepper } from '@/core/shared/ui/QuantityStepper';

import styles from './CatalogB2BTable.module.css';

type CatalogB2BTableProps = {
  onAddToCart: (productID: string, name: string, qty: number, packageQty: number) => void;
  products: ProductListItem[];
};

const ProductThumb = ({ product }: { product: ProductListItem }): JSX.Element => {
  const image = product.images?.find((item) => item.is_primary) ?? product.images?.[0];
  const [failed, setFailed] = useState(false);

  if (!image || failed) {
    return <ProductImageFallback categoryName={product.category_name} className={styles.thumb} />;
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

export const CatalogB2BTable = ({ onAddToCart, products }: CatalogB2BTableProps): JSX.Element => {
  const [qtyById, setQtyById] = useState<Record<string, number>>({});

  return (
    <div className={clsx(styles.scroll)}>
      <table className={clsx(styles.table)}>
        <thead>
          <tr>
            <th aria-label="Изображение" />
            <th>Артикул</th>
            <th>Наименование</th>
            <th>ГОСТ</th>
            <th>Материал</th>
            <th>Размер</th>
            <th>Уп.</th>
            <th>Остаток</th>
            <th>Цена</th>
            <th>Кол-во</th>
            <th aria-label="Действие" />
          </tr>
        </thead>
        <tbody>
          {products.map((product) => {
            const stockLevel = getStockLevel(product.stock_qty);
            const isOut = stockLevel === 'out';
            const qty = qtyById[product.id] ?? product.package_qty;
            const hasDiscount = product.discount_percent > 0;

            return (
              <tr key={product.id}>
                <td>
                  <div className={clsx(styles.thumb)}>
                    <ProductThumb product={product} />
                  </div>
                </td>
                <td className={clsx(styles.sku)}>{product.sku}</td>
                <td>
                  <Link className={clsx(styles.nameLink)} href={getProductPath(product.id)}>
                    {product.name}
                  </Link>
                </td>
                <td>{product.gost || '—'}</td>
                <td>{product.material || '—'}</td>
                <td>{product.size || '—'}</td>
                <td>{product.package_qty} шт</td>
                <td>
                  <span
                    className={clsx(
                      styles.stock,
                      stockLevel === 'out' && styles.stockOut,
                      stockLevel === 'low' && styles.stockLow,
                    )}
                  >
                    {product.stock_qty}
                  </span>
                </td>
                <td>
                  <span className={clsx(styles.price, hasDiscount && styles.pricePromo)}>
                    {formatPrice(product.client_price)}
                  </span>
                </td>
                <td>
                  <QuantityStepper
                    disabled={isOut}
                    onChange={(value) =>
                      setQtyById((current) => ({ ...current, [product.id]: value }))
                    }
                    step={product.package_qty}
                    value={qty}
                  />
                </td>
                <td>
                  <Button
                    className={clsx(styles.cartButton)}
                    isDisabled={isOut}
                    onPress={() => onAddToCart(product.id, product.name, qty, product.package_qty)}
                    size="sm"
                    variant="primary"
                  >
                    {isOut ? (
                      'Нет'
                    ) : (
                      <>
                        <IconPlus currentColor="currentColor" height={14} width={14} />В корзину
                      </>
                    )}{' '}
                  </Button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};
