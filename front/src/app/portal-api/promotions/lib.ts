import type {
  Promotion,
  PromotionDiscMode,
  PromotionSelection,
  PromotionType,
} from '@/core/shared/server/portal/types';

export type PromotionWriteBody = {
  condition?: string;
  desc?: string;
  discMode?: PromotionDiscMode;
  discValue?: number;
  endAt?: string;
  minQty?: number;
  name?: string;
  sel?: PromotionSelection;
  startAt?: string;
  type?: PromotionType;
};

const isDiscMode = (value: unknown): value is PromotionDiscMode =>
  value === 'percent' || value === 'price';

const isPromoType = (value: unknown): value is PromotionType => value === 'date' || value === 'qty';

const normalizeSel = (sel: PromotionSelection | undefined): PromotionSelection => {
  if (!sel || typeof sel !== 'object') {
    return { all: false, nodes: [], products: [] };
  }

  return {
    all: Boolean(sel.all),
    nodes: Array.isArray(sel.nodes) ? sel.nodes.map(String).filter(Boolean) : [],
    products: Array.isArray(sel.products) ? sel.products.map(String).filter(Boolean) : [],
  };
};

export const validatePromotionPayload = (
  body: null | PromotionWriteBody,
): { error: string } | { value: Omit<Promotion, 'createdAt' | 'endedManually' | 'id'> } => {
  if (!body) {
    return { error: 'Некорректное тело запроса' };
  }

  const name = body.name?.trim() ?? '';

  if (!name) {
    return { error: 'Укажите название акции' };
  }

  const startAt = body.startAt?.trim() ?? '';
  const endAt = body.endAt?.trim() ?? '';

  if (!startAt || !endAt) {
    return { error: 'Укажите период действия' };
  }

  if (new Date(endAt).getTime() <= new Date(startAt).getTime()) {
    return { error: 'Окончание должно быть позже начала' };
  }

  if (!isPromoType(body.type)) {
    return { error: 'Укажите тип акции' };
  }

  if (!isDiscMode(body.discMode)) {
    return { error: 'Укажите тип скидки' };
  }

  const discValue = Number(body.discValue);

  if (!Number.isFinite(discValue) || discValue <= 0) {
    return { error: 'Укажите размер скидки или цену' };
  }

  const minQty = body.type === 'qty' ? Math.max(1, Number(body.minQty) || 1) : 0;
  const sel = normalizeSel(body.sel);

  if (!sel.all && sel.nodes.length === 0 && sel.products.length === 0) {
    return { error: 'Выберите товары акции' };
  }

  return {
    value: {
      condition: body.condition?.trim() ?? '',
      desc: body.desc?.trim() ?? '',
      discMode: body.discMode,
      discValue,
      endAt,
      minQty,
      name,
      sel,
      startAt,
      type: body.type,
    },
  };
};
