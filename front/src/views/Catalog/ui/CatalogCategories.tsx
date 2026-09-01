'use client';

import clsx from 'clsx';
import { useMemo, useState } from 'react';

import type { Category } from '@/core/shared/api/products';

import { IconChevronRight } from '@/core/shared/icons/IconChevronRight';
import { IconFavorite } from '@/core/shared/icons/IconFavorite';
import { IconLayers } from '@/core/shared/icons/IconLayers';
import { IconOrders } from '@/core/shared/icons/IconOrders';

import type { CatalogCollection } from '../lib/filters';

import styles from './CatalogCategories.module.css';

type CatalogCategoriesProps = {
  categories: Category[];
  onSelect: (categoryID?: string) => void;
  onSelectCollection: (collection?: CatalogCollection) => void;
  selectedCategoryId?: string;
  selectedCollection?: CatalogCollection;
  totalAll: number;
};

type CategoryNode = Category & {
  children: CategoryNode[];
};

const buildTree = (categories: Category[]): CategoryNode[] => {
  const map = new Map<string, CategoryNode>();

  categories.forEach((category) => {
    map.set(category.id, { ...category, children: [] });
  });

  const roots: CategoryNode[] = [];

  map.forEach((node) => {
    if (node.parent_id && map.has(node.parent_id)) {
      map.get(node.parent_id)?.children.push(node);
    } else {
      roots.push(node);
    }
  });

  return roots;
};

export const CatalogCategories = ({
  categories,
  onSelect,
  onSelectCollection,
  selectedCategoryId,
  selectedCollection,
  totalAll,
}: CatalogCategoriesProps): JSX.Element => {
  const tree = useMemo(() => buildTree(categories), [categories]);
  const [expandedIds, setExpandedIds] = useState<string[]>([]);
  const isAllActive = !selectedCategoryId && !selectedCollection;

  const toggleExpanded = (id: string): void => {
    setExpandedIds((current) =>
      current.includes(id) ? current.filter((item) => item !== id) : [...current, id],
    );
  };

  const renderNode = (node: CategoryNode, depth = 0): JSX.Element => {
    const hasChildren = node.children.length > 0;
    const isExpanded = expandedIds.includes(node.id) || selectedCategoryId === node.id;
    const isActive = selectedCategoryId === node.id;

    return (
      <li key={node.id}>
        <div className={clsx(styles.row)} style={{ paddingInlineStart: depth > 0 ? 12 : 0 }}>
          <button
            className={clsx(styles.item, isActive && styles.itemActive)}
            onClick={() => {
              if (hasChildren) {
                toggleExpanded(node.id);
              }

              onSelect(node.id);
            }}
            type="button"
          >
            <span className={clsx(styles.itemLeading)}>
              <IconChevronRight
                className={clsx(
                  styles.chevron,
                  isExpanded && hasChildren && styles.chevronExpanded,
                )}
                height={14}
                width={14}
              />
              <span className={clsx(styles.itemLabel)}>{node.name}</span>
            </span>
          </button>
        </div>
        {hasChildren && isExpanded ? (
          <ul className={clsx(styles.list)}>
            {node.children.map((child) => renderNode(child, depth + 1))}
          </ul>
        ) : null}
      </li>
    );
  };

  return (
    <aside className={clsx(styles.root)}>
      <div className={clsx(styles.shortcuts)}>
        <button
          className={clsx(
            styles.shortcut,
            selectedCollection === 'favorites' && styles.shortcutActive,
          )}
          onClick={() => onSelectCollection('favorites')}
          type="button"
        >
          <IconFavorite currentColor="currentColor" height={16} width={16} />
          Моё избранное
        </button>
        <button
          className={clsx(
            styles.shortcut,
            selectedCollection === 'ordered' && styles.shortcutActive,
          )}
          onClick={() => onSelectCollection('ordered')}
          type="button"
        >
          <IconOrders currentColor="currentColor" height={16} width={16} />
          Ранее заказывал
        </button>
      </div>
      <h2 className={clsx(styles.title)}>Категории</h2>
      <ul className={clsx(styles.list)}>
        <li>
          <button
            className={clsx(styles.item, isAllActive && styles.itemActive)}
            onClick={() => onSelect(undefined)}
            type="button"
          >
            <span className={clsx(styles.itemLeading)}>
              <IconLayers
                className={clsx(styles.allIcon)}
                currentColor="currentColor"
                height={18}
                width={18}
              />
              <span className={clsx(styles.itemLabel)}>Все товары</span>
            </span>
            <span className={clsx(styles.count)}>{totalAll}</span>
          </button>
        </li>
        {tree.map((node) => renderNode(node))}
      </ul>
    </aside>
  );
};
