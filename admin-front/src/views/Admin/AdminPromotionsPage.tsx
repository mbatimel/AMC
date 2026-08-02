'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useCallback, useEffect, useState } from 'react';

import type { ProductListItem } from '@/core/shared/api/products';
import type {
  Promotion,
  PromotionStatus,
  PromotionWritePayload,
} from '@/core/shared/api/promotions';

import { $adminUserId } from '@/core/entities/adminSession';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { fetchAllProductsRequest } from '@/core/shared/api/products';
import {
  createPromotionRequest,
  deletePromotionRequest,
  fetchPromotionsRequest,
  updatePromotionRequest,
} from '@/core/shared/api/promotions';
import { FormSelect } from '@/core/shared/ui/FormSelect';

import styles from './Admin.module.css';
import { AdminPageHeader } from './ui/AdminPageHeader';

type PromotionDraft = {
  discountPercent: number;
  endsAt: string;
  id: string;
  name: string;
  products: { minQty: number; productId: string }[];
  startsAt: string;
  status: null | PromotionStatus;
};

const statusLabel: Record<PromotionStatus, string> = {
  active: 'Активна',
  ended: 'Завершена',
  scheduled: 'Запланирована',
};

const statusBadgeClass: Record<PromotionStatus, string> = {
  active: styles.badgeSuccess,
  ended: styles.badge,
  scheduled: styles.badgeWarning,
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
  name: 'Новая акция',
  products: [],
  startsAt: '',
  status: null,
});

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
  const [drafts, setDrafts] = useState<PromotionDraft[]>([]);
  const [error, setError] = useState<null | string>(null);
  const [itemErrors, setItemErrors] = useState<Record<string, string>>({});
  const [savingId, setSavingId] = useState<null | string>(null);
  const [isLoading, setIsLoading] = useState(true);

  const load = useCallback(async (): Promise<void> => {
    if (!adminUserId) {
      return;
    }
    try {
      const [promotions, allProducts] = await Promise.all([
        fetchPromotionsRequest(adminUserId),
        fetchAllProductsRequest(),
      ]);

      setDrafts(promotions.map(draftFromPromotion));
      setProducts(allProducts);
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

  const productName = (productId: string): string => {
    const product = products.find((item) => item.id === productId);

    return product ? `${product.name} (${product.sku})` : productId;
  };

  const patchDraft = (id: string, patch: Partial<PromotionDraft>): void => {
    setDrafts((previous) =>
      previous.map((draft) => (draft.id === id ? { ...draft, ...patch } : draft)),
    );
  };

  const addProductToDraft = (draftId: string, productId: string): void => {
    if (!productId) {
      return;
    }
    setDrafts((previous) =>
      previous.map((draft) =>
        draft.id === draftId && !draft.products.some((product) => product.productId === productId)
          ? { ...draft, products: [...draft.products, { minQty: 1, productId }] }
          : draft,
      ),
    );
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
          <button
            className={clsx(styles.smallButton)}
            onClick={() => setDrafts((previous) => [createEmptyDraft(), ...previous])}
            type="button"
          >
            Добавить акцию
          </button>
        }
        subtitle="Скидка в процентах на выбранные товары на период, от порога количества"
        title="Акции"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}
      {isLoading ? <p className={clsx(styles.hint)}>Загружаем акции…</p> : null}

      <div className={clsx(styles.listEditor)}>
        {drafts.map((draft) => {
          const isNew = draft.id.startsWith('new-');
          const availableProducts = products.filter(
            (product) => !draft.products.some((item) => item.productId === product.id),
          );

          return (
            <article className={clsx(styles.card)} key={draft.id}>
              <div className={clsx(styles.bannerHeader)}>
                <strong>
                  {draft.status ? statusLabel[draft.status] : 'Новая акция'}
                  {draft.status ? (
                    <span className={clsx(styles.badge, statusBadgeClass[draft.status])}> </span>
                  ) : null}
                </strong>
                <div className={clsx(styles.rowActions)}>
                  <button
                    className={clsx(styles.smallButton, styles.smallButtonDanger)}
                    onClick={() => void remove(draft)}
                    type="button"
                  >
                    Удалить
                  </button>
                </div>
              </div>

              {itemErrors[draft.id] ? (
                <p className={clsx(styles.error)}>{itemErrors[draft.id]}</p>
              ) : null}

              <div className={clsx(styles.formGrid)}>
                <label className={clsx(styles.field)}>
                  Название
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchDraft(draft.id, { name: event.target.value })}
                    value={draft.name}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  Скидка, %
                  <input
                    className={clsx(styles.input)}
                    max={100}
                    min={0}
                    onChange={(event) =>
                      patchDraft(draft.id, { discountPercent: Number(event.target.value) || 0 })
                    }
                    type="number"
                    value={draft.discountPercent}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  Начало акции
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchDraft(draft.id, { startsAt: event.target.value })}
                    type="datetime-local"
                    value={draft.startsAt}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  Конец акции
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchDraft(draft.id, { endsAt: event.target.value })}
                    type="datetime-local"
                    value={draft.endsAt}
                  />
                </label>
              </div>

              <h3 className={clsx(styles.cardTitle)}>Товары акции ({draft.products.length})</h3>
              <FormSelect
                ariaLabel="Добавить товар в акцию"
                onChange={(productId) => addProductToDraft(draft.id, productId)}
                options={availableProducts.map((product) => ({
                  label: `${product.name} (${product.sku})`,
                  value: product.id,
                }))}
                placeholder="Добавить товар…"
                value=""
              />
              <div className={clsx(styles.listEditor)}>
                {draft.products.map((product) => (
                  <div className={clsx(styles.listRow)} key={product.productId}>
                    <span>{productName(product.productId)}</span>
                    <label className={clsx(styles.field)}>
                      Мин. кол-во
                      <input
                        className={clsx(styles.input)}
                        min={1}
                        onChange={(event) =>
                          patchDraftProduct(
                            draft.id,
                            product.productId,
                            Math.max(1, Number(event.target.value) || 1),
                          )
                        }
                        type="number"
                        value={product.minQty}
                      />
                    </label>
                    <button
                      className={clsx(styles.smallButton, styles.smallButtonDanger)}
                      onClick={() => removeProductFromDraft(draft.id, product.productId)}
                      type="button"
                    >
                      Убрать
                    </button>
                  </div>
                ))}
                {draft.products.length === 0 ? (
                  <p className={clsx(styles.hint)}>Товары не выбраны</p>
                ) : null}
              </div>

              <Button
                isDisabled={savingId === draft.id}
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
