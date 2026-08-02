'use client';

import clsx from 'clsx';
import { useEffect, useRef } from 'react';

import type { PromotionSelection } from '@/core/shared/api/promotions';

import styles from '../Admin.module.css';
import promoStyles from '../AdminPromotions.module.css';
import {
  applySelectionToTree,
  collectSelectionFromTree,
  type PromoTreeNode,
  toggleTreeNode,
  toggleTreeProduct,
  type TreeCheckState,
} from '../lib/promotions';

type PromoProductTreeProps = {
  onChange: (sel: PromotionSelection) => void;
  roots: PromoTreeNode[];
  selection: PromotionSelection;
};

const TreeBranch = ({
  depth,
  disabled,
  node,
  onNodeToggle,
  onProductToggle,
  state,
}: {
  depth: number;
  disabled: boolean;
  node: PromoTreeNode;
  onNodeToggle: (nodeId: string, checked: boolean) => void;
  onProductToggle: (productId: string, checked: boolean) => void;
  state: TreeCheckState;
}): JSX.Element => {
  const nodeRef = useRef<HTMLInputElement>(null);
  const indeterminate = state.indeterminateNodes.has(node.id);

  useEffect(() => {
    if (nodeRef.current) {
      nodeRef.current.indeterminate = indeterminate;
    }
  }, [indeterminate]);

  return (
    <div className={clsx(promoStyles.treeNode)}>
      <label className={clsx(promoStyles.treeRow)} style={{ paddingInlineStart: depth * 18 }}>
        <input
          checked={state.checkedNodes.has(node.id)}
          disabled={disabled}
          onChange={(event) => onNodeToggle(node.id, event.target.checked)}
          ref={nodeRef}
          type="checkbox"
        />
        <span>{node.name}</span>
      </label>

      {node.children.map((child) => (
        <TreeBranch
          depth={depth + 1}
          disabled={disabled}
          key={child.id}
          node={child}
          onNodeToggle={onNodeToggle}
          onProductToggle={onProductToggle}
          state={state}
        />
      ))}

      {node.products.map((product) => (
        <label
          className={clsx(promoStyles.treeRow, promoStyles.treeProd)}
          key={product.id}
          style={{ paddingInlineStart: (depth + 1) * 18 }}
        >
          <input
            checked={state.checkedProducts.has(product.id)}
            disabled={disabled}
            onChange={(event) => onProductToggle(product.id, event.target.checked)}
            type="checkbox"
          />
          <span className={clsx(promoStyles.treeSku)}>{product.sku}</span>
          <span>{product.name}</span>
        </label>
      ))}
    </div>
  );
};

export const PromoProductTree = ({
  onChange,
  roots,
  selection,
}: PromoProductTreeProps): JSX.Element => {
  const all = selection.all;
  const state = applySelectionToTree(roots, selection);

  return (
    <div className={clsx(styles.field)}>
      <span className={clsx(styles.label)}>Товары акции *</span>
      <label className={clsx(styles.checkboxRow)}>
        <input
          checked={all}
          onChange={(event) => {
            const checked = event.target.checked;

            onChange(
              checked
                ? { all: true, nodes: [], products: [] }
                : { all: false, nodes: [], products: [] },
            );
          }}
          type="checkbox"
        />
        Все товары
      </label>

      <div className={clsx(promoStyles.promoTree, all && promoStyles.promoTreeDisabled)}>
        {roots.length === 0 ? (
          <p className={clsx(styles.hint)}>Категории и товары ещё не загружены</p>
        ) : (
          roots.map((node) => (
            <TreeBranch
              depth={0}
              disabled={all}
              key={node.id}
              node={node}
              onNodeToggle={(nodeId, checked) => {
                const next = toggleTreeNode(roots, state, nodeId, checked);

                onChange(collectSelectionFromTree(roots, next, false));
              }}
              onProductToggle={(productId, checked) => {
                const next = toggleTreeProduct(roots, state, productId, checked);

                onChange(collectSelectionFromTree(roots, next, false));
              }}
              state={state}
            />
          ))
        )}
      </div>
    </div>
  );
};
