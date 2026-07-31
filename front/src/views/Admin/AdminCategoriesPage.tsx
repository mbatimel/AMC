'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect } from 'react';

import type { Category } from '@/core/shared/api/products';

import { AppPath } from '@/core/shared/router/paths';

import styles from './Admin.module.css';
import { $adminBrands, $adminCategories, adminCatalogOpened } from './model/catalog';
import { AdminPageHeader } from './ui/AdminPageHeader';

type CategoryNode = Category & { children: CategoryNode[] };

const buildTree = (categories: Category[]): CategoryNode[] => {
  const nodes = new Map<string, CategoryNode>();

  categories.forEach((category) => {
    nodes.set(category.id, { ...category, children: [] });
  });

  const roots: CategoryNode[] = [];

  nodes.forEach((node) => {
    const parent = node.parent_id ? nodes.get(node.parent_id) : undefined;

    if (parent) {
      parent.children.push(node);
    } else {
      roots.push(node);
    }
  });

  return roots;
};

const CategoryTree = ({ nodes }: { nodes: CategoryNode[] }): JSX.Element => (
  <ul className={clsx(styles.tree)}>
    {nodes.map((node) => (
      <li key={node.id}>
        <div className={clsx(styles.treeNode)}>
          <Link href={`${AppPath.Catalog}?categoryID=${node.id}`}>{node.name}</Link>
        </div>
        {node.children.length > 0 ? (
          <div className={clsx(styles.treeChildren)}>
            <CategoryTree nodes={node.children} />
          </div>
        ) : null}
      </li>
    ))}
  </ul>
);

export const AdminCategoriesPage = (): JSX.Element => {
  const [categories, brands, open] = useUnit([
    $adminCategories,
    $adminBrands,
    adminCatalogOpened,
  ]);

  useEffect(() => {
    open();
  }, [open]);

  const tree = buildTree(categories);

  return (
    <>
      <AdminPageHeader
        subtitle="Дерево каталога и список брендов приходят из сервиса products (M-01) и синхронизируются с 1С"
        title="Категории и бренды"
      />

      <div className={clsx(styles.grid2)}>
        <section className={clsx(styles.card)}>
          <h2 className={clsx(styles.cardTitle)}>Категории ({categories.length})</h2>
          {tree.length === 0 ? (
            <p className={clsx(styles.hint)}>Категории не загружены</p>
          ) : (
            <CategoryTree nodes={tree} />
          )}
          <p className={clsx(styles.hint)}>
            Структура категорий ведётся в 1С. На портале доступен просмотр и привязка товара к
            категории в карточке товара.
          </p>
        </section>

        <section className={clsx(styles.card)}>
          <h2 className={clsx(styles.cardTitle)}>Бренды ({brands.length})</h2>
          {brands.length === 0 ? (
            <p className={clsx(styles.hint)}>Бренды не загружены</p>
          ) : (
            <ul className={clsx(styles.tree)}>
              {brands.map((brand) => (
                <li className={clsx(styles.treeNode)} key={brand.id}>
                  {brand.name}
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </>
  );
};
