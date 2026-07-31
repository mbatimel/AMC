'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect } from 'react';

import { formatPrice } from '@/core/shared/lib/formatPrice';
import { getAdminProductPath } from '@/core/shared/router/paths';

import styles from './Admin.module.css';
import {
  $adminCatalogError,
  $adminProducts,
  $adminProductQuery,
  $adminProductsTotal,
  $isAdminCatalogPending,
  adminCatalogOpened,
  adminProductDeleted,
  adminProductQueryChanged,
} from './model/catalog';
import { AdminPageHeader } from './ui/AdminPageHeader';

export const AdminProductsPage = (): JSX.Element => {
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

  return (
    <>
      <AdminPageHeader
        actions={
          <Link
            className={clsx(styles.smallButton, styles.smallButtonPrimary)}
            href={getAdminProductPath('new')}
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

            {products.map((product) => (
              <tr key={product.id}>
                <td>{product.sku}</td>
                <td>
                  <Link href={getAdminProductPath(product.id)}>{product.name}</Link>
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
                    <Link className={clsx(styles.smallButton)} href={getAdminProductPath(product.id)}>
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
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
};
