'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useMemo, useState } from 'react';

import type { Category } from '@/core/shared/api/products';

import styles from './CatalogCategories.module.css';

type CatalogCategoriesProps = {
  categories: Category[];
  onSelect: (categoryID?: string) => void;
  selectedCategoryId?: string;
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
  selectedCategoryId,
  totalAll,
}: CatalogCategoriesProps): JSX.Element => {
  const tree = useMemo(() => buildTree(categories), [categories]);
  const [expandedIds, setExpandedIds] = useState<string[]>([]);

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
        <div className={clsx(styles.row)} style={{ paddingInlineStart: `${12 + depth * 12}px` }}>
          {hasChildren ? (
            <Button
              aria-expanded={isExpanded}
              aria-label={isExpanded ? 'Свернуть' : 'Развернуть'}
              className={clsx(styles.chevron)}
              isIconOnly
              onPress={() => toggleExpanded(node.id)}
              size="sm"
              variant="ghost"
            >
              {isExpanded ? '▾' : '▸'}
            </Button>
          ) : (
            <span className={clsx(styles.chevronSpacer)} />
          )}
          <Button
            className={clsx(styles.item, isActive && styles.itemActive)}
            fullWidth
            onPress={() => onSelect(node.id)}
            variant={isActive ? 'secondary' : 'ghost'}
          >
            <span>{node.name}</span>
          </Button>
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
      <h2 className={clsx(styles.title)}>Категории</h2>
      <ul className={clsx(styles.list)}>
        <li>
          <Button
            className={clsx(styles.item, !selectedCategoryId && styles.itemActive)}
            fullWidth
            onPress={() => onSelect(undefined)}
            variant={!selectedCategoryId ? 'secondary' : 'ghost'}
          >
            <span>Все товары</span>
            <span className={clsx(styles.count)}>{totalAll}</span>
          </Button>
        </li>
        {tree.map((node) => renderNode(node))}
      </ul>
    </aside>
  );
};
