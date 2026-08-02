'use client';

import { Button, Chip, FieldError, Input, Label, NumberField, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useCallback, useEffect, useMemo, useState } from 'react';

import type { Category, ProductListItem } from '@/core/shared/api/products';
import type {
  Promotion,
  PromotionStatus,
  PromotionWritePayload,
} from '@/core/shared/api/promotions';

import { $adminUserId } from '@/core/entities/adminSession';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { fetchAllProductsRequest, listCategoriesRequest } from '@/core/shared/api/products';
import {
  createPromotionRequest,
  deletePromotionRequest,
  fetchPromotionsRequest,
  updatePromotionRequest,
} from '@/core/shared/api/promotions';
import { IconChevronDown } from '@/core/shared/icons';
import { DateTimeField } from '@/core/shared/ui/DateTimeField';
import { FormSelect, type FormSelectOptionGroup } from '@/core/shared/ui/FormSelect';

import styles from './Admin.module.css';
import { AdminPageHeader } from './ui/AdminPageHeader';

type CategoryGroupNode = FormSelectOptionGroup & {
  children: CategoryGroupNode[];
  options: NonNullable<FormSelectOptionGroup['options']>;
};

type DraftFieldErrors = {
  discountPercent?: string;
  endsAt?: string;
  name?: string;
  products?: string;
  startsAt?: string;
};

type PromotionDraft = {
  discountPercent: number;
  endsAt: string;
  id: string;
  name: string;
  products: { minQty: number; productId: string }[];
  startsAt: string;
  status: null | PromotionStatus;
};

const validateDraft = (draft: PromotionDraft): DraftFieldErrors => {
  const errors: DraftFieldErrors = {};

  if (!draft.name.trim()) {
    errors.name = 'Укажите название';
  }

  if (!Number.isFinite(draft.discountPercent)) {
    errors.discountPercent = 'Укажите скидку';
  } else if (draft.discountPercent < 0 || draft.discountPercent > 100) {
    errors.discountPercent = 'Скидка должна быть от 0 до 100';
  }

  if (!draft.startsAt.trim()) {
    errors.startsAt = 'Укажите дату начала';
  }

  if (!draft.endsAt.trim()) {
    errors.endsAt = 'Укажите дату окончания';
  }

  if (draft.startsAt.trim() && draft.endsAt.trim()) {
    const startsAt = new Date(draft.startsAt).getTime();
    const endsAt = new Date(draft.endsAt).getTime();

    if (Number.isNaN(startsAt)) {
      errors.startsAt = 'Некорректная дата начала';
    }

    if (Number.isNaN(endsAt)) {
      errors.endsAt = 'Некорректная дата окончания';
    } else if (!Number.isNaN(startsAt) && endsAt <= startsAt) {
      errors.endsAt = 'Дата окончания должна быть позже начала';
    }
  }

  if (draft.products.length === 0) {
    errors.products = 'Добавьте хотя бы один товар';
  }

  return errors;
};

const hasFieldErrors = (errors: DraftFieldErrors): boolean => Object.keys(errors).length > 0;

const buildProductCategoryGroups = (
  categories: Category[],
  products: ProductListItem[],
): FormSelectOptionGroup[] => {
  const nodes = new Map<string, CategoryGroupNode>();

  categories.forEach((category) => {
    nodes.set(category.id, {
      children: [],
      id: category.id,
      label: category.name,
      options: [],
    });
  });

  const roots: CategoryGroupNode[] = [];

  categories.forEach((category) => {
    const node = nodes.get(category.id);

    if (!node) {
      return;
    }

    if (category.parent_id && nodes.has(category.parent_id)) {
      nodes.get(category.parent_id)?.children.push(node);
      return;
    }

    roots.push(node);
  });

  const uncategorized: NonNullable<FormSelectOptionGroup['options']> = [];

  products.forEach((product) => {
    const option = {
      label: `${product.name} (${product.sku})`,
      value: product.id,
    };
    const node = product.category_id ? nodes.get(product.category_id) : undefined;

    if (node) {
      node.options.push(option);
      return;
    }

    uncategorized.push(option);
  });

  const prune = (items: CategoryGroupNode[]): FormSelectOptionGroup[] =>
    items.flatMap((item) => {
      const children = prune(item.children);

      if (item.options.length === 0 && children.length === 0) {
        return [];
      }

      return [
        {
          children: children.length > 0 ? children : undefined,
          id: item.id,
          label: item.label,
          options: item.options,
        },
      ];
    });

  const groups = prune(roots);

  if (uncategorized.length > 0) {
    groups.push({
      id: '__uncategorized__',
      label: 'Без категории',
      options: uncategorized,
    });
  }

  return groups;
};

const statusLabel: Record<PromotionStatus, string> = {
  active: 'Активна',
  ended: 'Завершена',
  scheduled: 'Запланирована',
};

const statusChipColor: Record<PromotionStatus, 'accent' | 'default' | 'success'> = {
  active: 'success',
  ended: 'default',
  scheduled: 'accent',
};

const pad = (value: number): string => String(value).padStart(2, '0');

const toLocalInput = (iso: string): string => {
  const date = new Date(iso);

  if (!iso || Number.isNaN(date.getTime())) {
    return '';
  }

  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

const fromLocalInput = (value: string): string => {
  const date = new Date(value);

  return value && !Number.isNaN(date.getTime()) ? date.toISOString() : '';
};

const draftFromPromotion = (promotion: Promotion): PromotionDraft => ({
  discountPercent: promotion.discount_percent,
  endsAt: toLocalInput(promotion.ends_at),
  id: promotion.id,
  name: promotion.name,
  products: promotion.products.map((product) => ({
    minQty: product.min_qty,
    productId: product.product_id,
  })),
  startsAt: toLocalInput(promotion.starts_at),
  status: promotion.status,
});

const createEmptyDraft = (): PromotionDraft => ({
  discountPercent: 10,
  endsAt: '',
  id: `new-${crypto.randomUUID()}`,
  name: '',
  products: [],
  startsAt: '',
  status: null,
});

const cloneDraft = (draft: PromotionDraft): PromotionDraft => ({
  ...draft,
  products: draft.products.map((product) => ({ ...product })),
});

const serializeDraft = (draft: PromotionDraft): string =>
  JSON.stringify({
    discountPercent: draft.discountPercent,
    endsAt: draft.endsAt,
    name: draft.name,
    products: [...draft.products]
      .map((product) => ({ minQty: product.minQty, productId: product.productId }))
      .sort((left, right) => left.productId.localeCompare(right.productId)),
    startsAt: draft.startsAt,
  });

const isDraftDirty = (draft: PromotionDraft, baseline: PromotionDraft | undefined): boolean => {
  if (!baseline) {
    return true;
  }

  return serializeDraft(draft) !== serializeDraft(baseline);
};

const draftToPayload = (draft: PromotionDraft): PromotionWritePayload => ({
  discountPercent: draft.discountPercent,
  endsAt: fromLocalInput(draft.endsAt),
  name: draft.name,
  products: draft.products.map((product) => ({
    min_qty: product.minQty,
    product_id: product.productId,
  })),
  startsAt: fromLocalInput(draft.startsAt),
});

export const AdminPromotionsPage = (): JSX.Element => {
  const adminUserId = useUnit($adminUserId);
  const [products, setProducts] = useState<ProductListItem[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [drafts, setDrafts] = useState<PromotionDraft[]>([]);
  const [baselines, setBaselines] = useState<Record<string, PromotionDraft>>({});
  const [error, setError] = useState<null | string>(null);
  const [itemErrors, setItemErrors] = useState<Record<string, string>>({});
  const [fieldErrors, setFieldErrors] = useState<Record<string, DraftFieldErrors>>({});
  const [collapsedProducts, setCollapsedProducts] = useState<Record<string, boolean>>({});
  const [savingId, setSavingId] = useState<null | string>(null);
  const [isLoading, setIsLoading] = useState(true);

  const isProductsCollapsed = (draftId: string): boolean =>
    draftId.startsWith('new-')
      ? collapsedProducts[draftId] === true
      : collapsedProducts[draftId] !== false;

  const toggleProductsCollapsed = (draftId: string): void => {
    setCollapsedProducts((previous) => {
      const currentlyCollapsed = draftId.startsWith('new-')
        ? previous[draftId] === true
        : previous[draftId] !== false;

      return { ...previous, [draftId]: !currentlyCollapsed };
    });
  };

  const productGroups = useMemo(
    () => buildProductCategoryGroups(categories, products),
    [categories, products],
  );

  const load = useCallback(async (): Promise<void> => {
    if (!adminUserId) {
      return;
    }
    try {
      const [promotions, allProducts, allCategories] = await Promise.all([
        fetchPromotionsRequest(adminUserId),
        fetchAllProductsRequest(),
        listCategoriesRequest(),
      ]);

      const nextDrafts = promotions.map(draftFromPromotion);

      setDrafts(nextDrafts);
      setBaselines(
        Object.fromEntries(nextDrafts.map((draft) => [draft.id, cloneDraft(draft)] as const)),
      );
      setProducts(allProducts);
      setCategories(allCategories);
      setError(null);
    } catch (loadError) {
      setError(toDisplayErrorMessage(loadError, 'Не удалось загрузить акции'));
    } finally {
      setIsLoading(false);
    }
  }, [adminUserId]);

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 0);

    return () => window.clearTimeout(timeout);
  }, [load]);

  const resolveProduct = (productId: string): { brandName?: string; name: string; sku: string } => {
    const product = products.find((item) => item.id === productId);

    if (!product) {
      return { name: productId, sku: '' };
    }

    return {
      brandName: product.brand_name,
      name: product.name,
      sku: product.sku,
    };
  };

  const patchDraft = (id: string, patch: Partial<PromotionDraft>): void => {
    setDrafts((previous) =>
      previous.map((draft) => (draft.id === id ? { ...draft, ...patch } : draft)),
    );
    setFieldErrors((previous) => {
      const current = previous[id];

      if (!current) {
        return previous;
      }

      const next = { ...current };
      (Object.keys(patch) as (keyof PromotionDraft)[]).forEach((key) => {
        if (key === 'name') {
          delete next.name;
        }
        if (key === 'discountPercent') {
          delete next.discountPercent;
        }
        if (key === 'startsAt') {
          delete next.startsAt;
        }
        if (key === 'endsAt') {
          delete next.endsAt;
        }
        if (key === 'products') {
          delete next.products;
        }
      });

      return { ...previous, [id]: next };
    });
  };

  const syncDraftProducts = (draftId: string, productIds: string[]): void => {
    setDrafts((previous) =>
      previous.map((draft) => {
        if (draft.id !== draftId) {
          return draft;
        }

        const existing = new Map(
          draft.products.map((product) => [product.productId, product] as const),
        );

        return {
          ...draft,
          products: productIds.map(
            (productId) => existing.get(productId) ?? { minQty: 1, productId },
          ),
        };
      }),
    );
    setFieldErrors((previous) => {
      const current = previous[draftId];

      if (!current?.products) {
        return previous;
      }

      const next = { ...current };
      delete next.products;

      return { ...previous, [draftId]: next };
    });
  };

  const patchDraftProduct = (draftId: string, productId: string, minQty: number): void => {
    setDrafts((previous) =>
      previous.map((draft) =>
        draft.id === draftId
          ? {
              ...draft,
              products: draft.products.map((product) =>
                product.productId === productId ? { ...product, minQty } : product,
              ),
            }
          : draft,
      ),
    );
  };

  const removeProductFromDraft = (draftId: string, productId: string): void => {
    setDrafts((previous) =>
      previous.map((draft) =>
        draft.id === draftId
          ? {
              ...draft,
              products: draft.products.filter((product) => product.productId !== productId),
            }
          : draft,
      ),
    );
  };

  const save = async (draft: PromotionDraft): Promise<void> => {
    if (!adminUserId) {
      return;
    }

    const errors = validateDraft(draft);

    if (hasFieldErrors(errors)) {
      setFieldErrors((previous) => ({ ...previous, [draft.id]: errors }));
      setItemErrors((previous) => ({
        ...previous,
        [draft.id]: 'Заполните обязательные поля',
      }));
      if (errors.products) {
        setCollapsedProducts((previous) => ({ ...previous, [draft.id]: false }));
      }
      return;
    }

    setFieldErrors((previous) => ({ ...previous, [draft.id]: {} }));
    setItemErrors((previous) => ({ ...previous, [draft.id]: '' }));
    setSavingId(draft.id);
    try {
      const payload = draftToPayload(draft);
      const isNew = draft.id.startsWith('new-');

      if (isNew) {
        await createPromotionRequest(adminUserId, payload);
      } else {
        await updatePromotionRequest(adminUserId, draft.id, payload);
      }
      await load();
    } catch (saveError) {
      setItemErrors((previous) => ({
        ...previous,
        [draft.id]: toDisplayErrorMessage(saveError, 'Не удалось сохранить акцию'),
      }));
    } finally {
      setSavingId(null);
    }
  };

  const remove = async (draft: PromotionDraft): Promise<void> => {
    if (!adminUserId || draft.id.startsWith('new-')) {
      setDrafts((previous) => previous.filter((item) => item.id !== draft.id));
      setBaselines((previous) => {
        const next = { ...previous };
        delete next[draft.id];
        return next;
      });
      return;
    }
    if (!window.confirm(`Удалить акцию «${draft.name}»?`)) {
      return;
    }
    setSavingId(draft.id);
    try {
      await deletePromotionRequest(adminUserId, draft.id);
      await load();
    } catch (deleteError) {
      setItemErrors((previous) => ({
        ...previous,
        [draft.id]: toDisplayErrorMessage(deleteError, 'Не удалось удалить акцию'),
      }));
    } finally {
      setSavingId(null);
    }
  };

  return (
    <>
      <AdminPageHeader
        actions={
          <Button
            onPress={() => {
              const draft = createEmptyDraft();
              setDrafts((previous) => [draft, ...previous]);
              setBaselines((previous) => ({ ...previous, [draft.id]: cloneDraft(draft) }));
            }}
            size="sm"
            variant="secondary"
          >
            Добавить акцию
          </Button>
        }
        subtitle="Скидка в процентах на выбранные товары на период, от порога количества"
        title="Акции"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}
      {isLoading ? <p className={clsx(styles.hint)}>Загружаем акции…</p> : null}

      <div className={clsx(styles.listEditor)}>
        {drafts.map((draft) => {
          const isNew = draft.id.startsWith('new-');
          const errors = fieldErrors[draft.id] ?? {};
          const canSave = isDraftDirty(draft, baselines[draft.id]);

          return (
            <article className={clsx(styles.card)} key={draft.id}>
              <div className={clsx(styles.bannerHeader)}>
                <div className={clsx(styles.promoTitleRow)}>
                  <strong>{isNew ? 'Новая акция' : draft.name}</strong>
                  {draft.status ? (
                    <Chip color={statusChipColor[draft.status]} size="sm" variant="soft">
                      <Chip.Label>{statusLabel[draft.status]}</Chip.Label>
                    </Chip>
                  ) : null}
                </div>
                <div className={clsx(styles.rowActions)}>
                  <Button onPress={() => void remove(draft)} size="sm" variant="danger-soft">
                    Удалить
                  </Button>
                </div>
              </div>

              {itemErrors[draft.id] ? (
                <p className={clsx(styles.error)}>{itemErrors[draft.id]}</p>
              ) : null}

              <div className={clsx(styles.formGrid)}>
                <TextField
                  className={clsx(styles.field)}
                  isInvalid={Boolean(errors.name)}
                  isRequired
                  onChange={(name) => patchDraft(draft.id, { name })}
                  value={draft.name}
                >
                  <Label className={clsx(styles.label)}>Название</Label>
                  <Input
                    className={clsx(styles.fieldControl, styles.fieldControlInput)}
                    placeholder="Например: Скидка на крепёж"
                  />
                  {errors.name ? <FieldError>{errors.name}</FieldError> : null}
                </TextField>

                <NumberField
                  className={clsx(styles.field)}
                  isInvalid={Boolean(errors.discountPercent)}
                  isRequired
                  maxValue={100}
                  minValue={0}
                  onChange={(next) =>
                    patchDraft(draft.id, {
                      discountPercent:
                        typeof next === 'number' && Number.isFinite(next) ? next : Number.NaN,
                    })
                  }
                  value={Number.isFinite(draft.discountPercent) ? draft.discountPercent : undefined}
                >
                  <Label className={clsx(styles.label)}>Скидка, %</Label>
                  <NumberField.Group className={clsx(styles.numberGroup, styles.fieldControl)}>
                    <NumberField.Input className={clsx(styles.numberInput)} />
                  </NumberField.Group>
                  {errors.discountPercent ? (
                    <FieldError>{errors.discountPercent}</FieldError>
                  ) : null}
                </NumberField>

                <DateTimeField
                  error={errors.startsAt}
                  label="Начало акции"
                  onChange={(startsAt) => patchDraft(draft.id, { startsAt })}
                  value={draft.startsAt}
                />
                <DateTimeField
                  error={errors.endsAt}
                  label="Конец акции"
                  onChange={(endsAt) => patchDraft(draft.id, { endsAt })}
                  value={draft.endsAt}
                />
              </div>

              <section className={clsx(styles.productsSection)}>
                <button
                  aria-expanded={!isProductsCollapsed(draft.id)}
                  className={clsx(styles.productsHeader)}
                  onClick={() => toggleProductsCollapsed(draft.id)}
                  type="button"
                >
                  <IconChevronDown
                    className={clsx(
                      styles.productsChevron,
                      !isProductsCollapsed(draft.id) && styles.productsChevronOpen,
                    )}
                    height={14}
                    width={14}
                  />
                  <h3 className={clsx(styles.productsTitle)}>Товары акции</h3>
                  <span className={clsx(styles.productsCount)}>{draft.products.length}</span>
                </button>

                {!isProductsCollapsed(draft.id) ? (
                  <>
                    <FormSelect
                      ariaLabel="Товары акции"
                      error={errors.products}
                      groups={productGroups}
                      label="Выбор товаров"
                      onChange={(productIds) => syncDraftProducts(draft.id, productIds)}
                      placeholder="Начните вводить название или артикул…"
                      selectionMode="multiple"
                      value={draft.products.map((product) => product.productId)}
                    />

                    {draft.products.length > 0 ? (
                      <ul className={clsx(styles.productList)}>
                        {draft.products.map((product) => {
                          const meta = resolveProduct(product.productId);

                          return (
                            <li className={clsx(styles.productRow)} key={product.productId}>
                              <div className={clsx(styles.productMeta)}>
                                <span className={clsx(styles.productName)}>{meta.name}</span>
                                <span className={clsx(styles.productSku)}>
                                  {[meta.sku ? `Арт. ${meta.sku}` : null, meta.brandName]
                                    .filter(Boolean)
                                    .join(' · ')}
                                </span>
                              </div>

                              <div className={clsx(styles.productControls)}>
                                <NumberField
                                  aria-label={`Мин. кол-во для ${meta.name}`}
                                  className={clsx(styles.minQtyField)}
                                  minValue={1}
                                  onChange={(next) =>
                                    patchDraftProduct(
                                      draft.id,
                                      product.productId,
                                      typeof next === 'number' && Number.isFinite(next)
                                        ? Math.max(1, next)
                                        : 1,
                                    )
                                  }
                                  value={product.minQty}
                                >
                                  <Label className={clsx(styles.minQtyLabel)}>От, шт.</Label>
                                  <NumberField.Group
                                    className={clsx(styles.numberGroup, styles.fieldControl)}
                                  >
                                    <NumberField.DecrementButton aria-label="Уменьшить мин. кол-во" />
                                    <NumberField.Input className={clsx(styles.numberInput)} />
                                    <NumberField.IncrementButton aria-label="Увеличить мин. кол-во" />
                                  </NumberField.Group>
                                </NumberField>

                                <Button
                                  aria-label={`Убрать ${meta.name}`}
                                  className={clsx(styles.removeProductButton)}
                                  isIconOnly
                                  onPress={() =>
                                    removeProductFromDraft(draft.id, product.productId)
                                  }
                                  size="sm"
                                  variant="danger-soft"
                                >
                                  ×
                                </Button>
                              </div>
                            </li>
                          );
                        })}
                      </ul>
                    ) : (
                      <p className={clsx(styles.productsEmpty)}>
                        Пока пусто — выберите один или несколько товаров выше
                      </p>
                    )}
                  </>
                ) : null}
              </section>

              <Button
                isDisabled={savingId === draft.id || !canSave}
                onPress={() => void save(draft)}
                variant="primary"
              >
                {savingId === draft.id ? 'Сохраняем…' : isNew ? 'Создать' : 'Сохранить'}
              </Button>
            </article>
          );
        })}
        {!isLoading && drafts.length === 0 ? (
          <p className={clsx(styles.empty)}>Акций пока нет</p>
        ) : null}
      </div>
    </>
  );
};
