'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect, useState } from 'react';

import type { ProductImageBatchResult } from '@/core/shared/api/products';

import { $adminUserId } from '@/core/entities/adminSession';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { uploadProductImagesBatchRequest } from '@/core/shared/api/products';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { getPrimaryProductImageUrl } from '@/core/shared/lib/productImage';
import { getProductPath } from '@/core/shared/router/paths';

import styles from './Admin.module.css';
import {
  $adminCatalogError,
  $adminProductQuery,
  $adminProducts,
  $adminProductsTotal,
  $isAdminCatalogPending,
  adminCatalogOpened,
  adminProductDeleted,
  adminProductQueryChanged,
} from './model/catalog';
import { AdminPageHeader } from './ui/AdminPageHeader';

export const AdminProductsPage = (): JSX.Element => {
  const adminUserId = useUnit($adminUserId);
  const [batchResults, setBatchResults] = useState<ProductImageBatchResult[]>([]);
  const [batchError, setBatchError] = useState<null | string>(null);
  const [isBatchPending, setIsBatchPending] = useState(false);
  const [products, total, query, isPending, error, open, changeQuery, remove] = useUnit([
    $adminProducts,
    $adminProductsTotal,
    $adminProductQuery,
    $isAdminCatalogPending,
    $adminCatalogError,
    adminCatalogOpened,
    adminProductQueryChanged,
    adminProductDeleted,
  ]);

  useEffect(() => {
    open();
  }, [open]);

  const uploadBatch = async (files: File[]): Promise<void> => {
    if (!adminUserId || files.length === 0) {
      return;
    }
    setBatchError(null);
    setIsBatchPending(true);
    try {
      setBatchResults(await uploadProductImagesBatchRequest(adminUserId, files));
      open();
    } catch (uploadError) {
      setBatchError(toDisplayErrorMessage(uploadError, 'Не удалось загрузить фотографии'));
    } finally {
      setIsBatchPending(false);
    }
  };

  return (
    <>
      <AdminPageHeader
        actions={
          <Link
            className={clsx(styles.smallButton, styles.smallButtonPrimary)}
            href={getProductPath('new')}
          >
            Добавить товар
          </Link>
        }
        subtitle={`${total} позиций в каталоге`}
        title="Товары"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <section className={clsx(styles.card)}>
        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="admin-product-search">
            Поиск по названию, артикулу или ГОСТ
          </label>
          <input
            className={clsx(styles.input)}
            id="admin-product-search"
            onChange={(event) => changeQuery(event.target.value)}
            placeholder="Например: метчик М8"
            value={query}
          />
        </div>
        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="product-images-batch">
            Массовая загрузка фото по SKU в имени файла
          </label>
          <input
            accept="image/jpeg,image/png,image/webp"
            disabled={isBatchPending}
            id="product-images-batch"
            multiple
            onChange={(event) => {
              void uploadBatch(Array.from(event.target.files ?? []));
              event.target.value = '';
            }}
            type="file"
          />
          {batchError ? <p className={clsx(styles.error)}>{batchError}</p> : null}
          {batchResults.map((result, index) => (
            <p
              className={clsx(result.success ? styles.hint : styles.error)}
              key={`${result.fileName}-${index}`}
            >
              {result.fileName} —{' '}
              {result.success ? `загружено для ${result.sku}` : result.errorText}
            </p>
          ))}
        </div>
      </section>

      <div className={clsx(styles.tableWrap)}>
        <table className={clsx(styles.table)}>
          <thead>
            <tr>
              <th>Артикул</th>
              <th>Наименование</th>
              <th>ГОСТ</th>
              <th>Остаток</th>
              <th>Цена</th>
              <th>Фото</th>
              <th>Статус</th>
              <th aria-label="Действия" />
            </tr>
          </thead>
          <tbody>
            {isPending && products.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={8}>
                  Загружаем каталог…
                </td>
              </tr>
            ) : null}

            {!isPending && products.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={8}>
                  Товары не найдены
                </td>
              </tr>
            ) : null}

            {products.map((product) => {
              const imageUrl = getPrimaryProductImageUrl(product.images);

              return (
                <tr key={product.id}>
                  <td>
                    <div className={clsx(styles.productThumbCell)}>
                      {imageUrl ? (
                        <img
                          alt={product.name}
                          className={clsx(styles.productThumb)}
                          src={imageUrl}
                        />
                      ) : (
                        <span className={clsx(styles.productThumbEmpty)}>—</span>
                      )}
                      <span>{product.sku}</span>
                    </div>
                  </td>
                  <td>
                    <Link href={getProductPath(product.id)}>{product.name}</Link>
                  </td>
                  <td>{product.gost || '—'}</td>
                  <td>{product.stock_qty}</td>
                  <td>{formatPrice(product.base_price)}</td>
                  <td>
                    {(product.images ?? []).length > 0 ? (
                      <span className={clsx(styles.badge, styles.badgeSuccess)}>есть</span>
                    ) : (
                      <span className={clsx(styles.badge, styles.badgeWarning)}>нет</span>
                    )}
                  </td>
                  <td>
                    {product.is_published === false ? (
                      <span className={clsx(styles.badge)}>Скрыт</span>
                    ) : (
                      <span className={clsx(styles.badge, styles.badgeSuccess)}>Опубликован</span>
                    )}
                  </td>
                  <td>
                    <div className={clsx(styles.rowActions)}>
                      <Link className={clsx(styles.smallButton)} href={getProductPath(product.id)}>
                        Изменить
                      </Link>
                      <button
                        className={clsx(styles.smallButton, styles.smallButtonDanger)}
                        onClick={() => remove(product.id)}
                        type="button"
                      >
                        Удалить
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </>
  );
};
