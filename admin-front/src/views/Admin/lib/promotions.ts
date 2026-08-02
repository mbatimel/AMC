import type { Category, ProductListItem } from '@/core/shared/api/products';
import type { Promotion, PromotionSelection, PromotionType } from '@/core/shared/api/promotions';

export type PromotionStatus = 'active' | 'ended' | 'scheduled';

export type PromoTreeNode = Category & {
  children: PromoTreeNode[];
  products: ProductListItem[];
};

export const PROMOTION_TYPE_LABELS: Record<PromotionType, string> = {
  date: 'Без доп. условий',
  qty: 'От количества товара',
};

export const PROMOTION_STATUS_LABELS: Record<PromotionStatus, string> = {
  active: 'Идёт',
  ended: 'Завершена',
  scheduled: 'Запланирована',
};

export const getPromotionStatus = (promo: Promotion, now = Date.now()): PromotionStatus => {
  if (promo.endedManually) {
    return 'ended';
  }

  const start = promo.startAt ? new Date(promo.startAt).getTime() : -Infinity;
  const end = promo.endAt ? new Date(promo.endAt).getTime() : Infinity;

  if (now < start) {
    return 'scheduled';
  }

  if (now > end) {
    return 'ended';
  }

  return 'active';
};

export const formatPromotionDiscount = (promo: Promotion): string => {
  if (promo.discMode === 'price') {
    return new Intl.NumberFormat('ru-RU', {
      currency: 'RUB',
      maximumFractionDigits: 0,
      style: 'currency',
    }).format(Math.round(promo.discValue));
  }

  return `−${promo.discValue || 0}%`;
};

export const toDatetimeLocalValue = (value: string): string => {
  if (!value) {
    return '';
  }

  if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(value)) {
    return value;
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  const pad = (part: number): string => String(part).padStart(2, '0');

  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

const declRus = (n: number, forms: [string, string, string]): string => {
  const abs = Math.abs(n) % 100;
  const last = abs % 10;

  if (abs > 10 && abs < 20) {
    return forms[2];
  }

  if (last > 1 && last < 5) {
    return forms[1];
  }

  if (last === 1) {
    return forms[0];
  }

  return forms[2];
};

const categoryDescendants = (roots: PromoTreeNode[], id: string): Set<string> => {
  const set = new Set<string>();
  let target: null | PromoTreeNode = null;

  const find = (nodes: PromoTreeNode[]): void => {
    for (const node of nodes) {
      if (node.id === id) {
        target = node;

        return;
      }

      find(node.children);
    }
  };

  find(roots);

  if (!target) {
    set.add(id);

    return set;
  }

  const collect = (node: PromoTreeNode): void => {
    set.add(node.id);
    node.children.forEach(collect);
  };

  collect(target);

  return set;
};

export const productMatchesSelection = (
  product: ProductListItem,
  sel: PromotionSelection,
  roots: PromoTreeNode[],
): boolean => {
  if (sel.all) {
    return true;
  }

  if (sel.products.includes(product.id)) {
    return true;
  }

  for (const nodeId of sel.nodes) {
    const descendants = categoryDescendants(roots, nodeId);

    if (product.category_id && descendants.has(product.category_id)) {
      return true;
    }
  }

  return false;
};

export const summarizePromotionSelection = (
  promo: Promotion,
  roots: PromoTreeNode[],
  products: ProductListItem[],
): string => {
  const sel = promo.sel;

  if (sel.all) {
    return 'Все товары';
  }

  const categoryNames = new Map<string, string>();

  const walk = (nodes: PromoTreeNode[]): void => {
    nodes.forEach((node) => {
      categoryNames.set(node.id, node.name);
      walk(node.children);
    });
  };

  walk(roots);

  const parts: string[] = [];

  sel.nodes.forEach((id) => {
    const name = categoryNames.get(id);

    if (name) {
      parts.push(name);
    }
  });

  sel.products.forEach((id) => {
    const product = products.find((item) => item.id === id);

    if (product) {
      parts.push(product.sku);
    }
  });

  const total = products.filter((product) => productMatchesSelection(product, sel, roots)).length;

  return `${parts.length ? parts.join(', ') : '—'} · ${total} ${declRus(total, ['товар', 'товара', 'товаров'])}`;
};

export const buildPromoTree = (
  categories: Category[],
  products: ProductListItem[],
): PromoTreeNode[] => {
  const nodes = new Map<string, PromoTreeNode>();

  categories.forEach((category) => {
    nodes.set(category.id, { ...category, children: [], products: [] });
  });

  const roots: PromoTreeNode[] = [];

  nodes.forEach((node) => {
    const parent = node.parent_id ? nodes.get(node.parent_id) : undefined;

    if (parent) {
      parent.children.push(node);
    } else {
      roots.push(node);
    }
  });

  products.forEach((product) => {
    if (!product.category_id) {
      return;
    }

    const node = nodes.get(product.category_id);

    if (node) {
      node.products.push(product);
    }
  });

  return roots;
};

export type TreeCheckState = {
  checkedNodes: Set<string>;
  checkedProducts: Set<string>;
  indeterminateNodes: Set<string>;
};

const emptyCheckState = (): TreeCheckState => ({
  checkedNodes: new Set(),
  checkedProducts: new Set(),
  indeterminateNodes: new Set(),
});

const collectDescendantIds = (node: PromoTreeNode): { nodeIds: string[]; productIds: string[] } => {
  const nodeIds = [node.id];
  const productIds = node.products.map((product) => product.id);

  node.children.forEach((child) => {
    const nested = collectDescendantIds(child);

    nodeIds.push(...nested.nodeIds);
    productIds.push(...nested.productIds);
  });

  return { nodeIds, productIds };
};

export const applySelectionToTree = (
  roots: PromoTreeNode[],
  sel: PromotionSelection,
): TreeCheckState => {
  const state = emptyCheckState();

  if (sel.all) {
    return state;
  }

  const markNodeChecked = (nodeId: string): void => {
    const find = (nodes: PromoTreeNode[]): null | PromoTreeNode => {
      for (const node of nodes) {
        if (node.id === nodeId) {
          return node;
        }

        const nested = find(node.children);

        if (nested) {
          return nested;
        }
      }

      return null;
    };

    const target = find(roots);

    if (!target) {
      return;
    }

    const { nodeIds, productIds } = collectDescendantIds(target);

    nodeIds.forEach((id) => state.checkedNodes.add(id));
    productIds.forEach((id) => state.checkedProducts.add(id));
  };

  sel.nodes.forEach(markNodeChecked);
  sel.products.forEach((id) => state.checkedProducts.add(id));

  return recalcTreeChecks(roots, state);
};

export const recalcTreeChecks = (roots: PromoTreeNode[], draft: TreeCheckState): TreeCheckState => {
  const checkedNodes = new Set<string>();
  const indeterminateNodes = new Set<string>();
  const checkedProducts = new Set(draft.checkedProducts);

  const calc = (node: PromoTreeNode): { checked: boolean; indeterminate: boolean } => {
    const childStates = node.children.map(calc);
    const productStates = node.products.map((product) => ({
      checked: checkedProducts.has(product.id),
      indeterminate: false,
    }));
    const states = [...childStates, ...productStates];

    if (states.length === 0) {
      const checked = draft.checkedNodes.has(node.id);

      if (checked) {
        checkedNodes.add(node.id);
      }

      return { checked, indeterminate: false };
    }

    const allChecked = states.every((item) => item.checked && !item.indeterminate);
    const noneChecked = states.every((item) => !item.checked && !item.indeterminate);
    const indeterminate = !allChecked && !noneChecked;

    if (allChecked) {
      checkedNodes.add(node.id);
    }

    if (indeterminate) {
      indeterminateNodes.add(node.id);
    }

    return { checked: allChecked, indeterminate };
  };

  roots.forEach(calc);

  return { checkedNodes, checkedProducts, indeterminateNodes };
};

export const toggleTreeNode = (
  roots: PromoTreeNode[],
  state: TreeCheckState,
  nodeId: string,
  checked: boolean,
): TreeCheckState => {
  const find = (nodes: PromoTreeNode[]): null | PromoTreeNode => {
    for (const node of nodes) {
      if (node.id === nodeId) {
        return node;
      }

      const nested = find(node.children);

      if (nested) {
        return nested;
      }
    }

    return null;
  };

  const target = find(roots);

  if (!target) {
    return state;
  }

  const { nodeIds, productIds } = collectDescendantIds(target);
  const nextNodes = new Set(state.checkedNodes);
  const nextProducts = new Set(state.checkedProducts);

  nodeIds.forEach((id) => {
    if (checked) {
      nextNodes.add(id);
    } else {
      nextNodes.delete(id);
    }
  });

  productIds.forEach((id) => {
    if (checked) {
      nextProducts.add(id);
    } else {
      nextProducts.delete(id);
    }
  });

  return recalcTreeChecks(roots, {
    checkedNodes: nextNodes,
    checkedProducts: nextProducts,
    indeterminateNodes: new Set(),
  });
};

export const toggleTreeProduct = (
  roots: PromoTreeNode[],
  state: TreeCheckState,
  productId: string,
  checked: boolean,
): TreeCheckState => {
  const nextProducts = new Set(state.checkedProducts);

  if (checked) {
    nextProducts.add(productId);
  } else {
    nextProducts.delete(productId);
  }

  return recalcTreeChecks(roots, {
    checkedNodes: state.checkedNodes,
    checkedProducts: nextProducts,
    indeterminateNodes: new Set(),
  });
};

export const collectSelectionFromTree = (
  roots: PromoTreeNode[],
  state: TreeCheckState,
  all: boolean,
): PromotionSelection => {
  if (all) {
    return { all: true, nodes: [], products: [] };
  }

  const nodes: string[] = [];
  const products: string[] = [];

  const gather = (node: PromoTreeNode): void => {
    if (state.checkedNodes.has(node.id) && !state.indeterminateNodes.has(node.id)) {
      nodes.push(node.id);

      return;
    }

    if (!state.checkedNodes.has(node.id) && !state.indeterminateNodes.has(node.id)) {
      return;
    }

    node.children.forEach(gather);
    node.products.forEach((product) => {
      if (state.checkedProducts.has(product.id)) {
        products.push(product.id);
      }
    });
  };

  roots.forEach(gather);

  return { all: false, nodes, products };
};
