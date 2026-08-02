'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect, useMemo, useState } from 'react';

import type {
  Promotion,
  PromotionDiscMode,
  PromotionSelection,
  PromotionType,
  PromotionWritePayload,
} from '@/core/shared/api/promotions';

import styles from './Admin.module.css';
import promoStyles from './AdminPromotions.module.css';
import { formatAdminDateTime } from './lib/nav';
import {
  buildPromoTree,
  formatPromotionDiscount,
  getPromotionStatus,
  PROMOTION_STATUS_LABELS,
  PROMOTION_TYPE_LABELS,
  summarizePromotionSelection,
  toDatetimeLocalValue,
} from './lib/promotions';
import {
  $isPromoCatalogPending,
  $isPromotionSaving,
  $isPromotionsPending,
  $promoCategories,
  $promoProducts,
  $promotions,
  $promotionsError,
  adminPromotionsOpened,
  promotionDeleteRequested,
  promotionEndRequested,
  promotionSaveRequested,
} from './model/promotions';
import { AdminPageHeader } from './ui/AdminPageHeader';
import { PromoProductTree } from './ui/PromoProductTree';

type ConfirmState = null | { kind: 'delete'; promo: Promotion } | { kind: 'end'; promo: Promotion };

type EditorDraft = {
  condition: string;
  desc: string;
  discMode: PromotionDiscMode;
  discValue: string;
  endAt: string;
  id: null | string;
  minQty: string;
  name: string;
  sel: PromotionSelection;
  startAt: string;
  type: PromotionType;
};

const emptyDraft = (): EditorDraft => ({
  condition: '',
  desc: '',
  discMode: 'percent',
  discValue: '10',
  endAt: '',
  id: null,
  minQty: '3',
  name: '',
  sel: { all: false, nodes: [], products: [] },
  startAt: '',
  type: 'date',
});

const draftFromPromotion = (promo: Promotion): EditorDraft => ({
  condition: promo.condition,
  desc: promo.desc,
  discMode: promo.discMode,
  discValue: String(promo.discValue),
  endAt: toDatetimeLocalValue(promo.endAt),
  id: promo.id,
  minQty: String(promo.minQty || 3),
  name: promo.name,
  sel: {
    all: Boolean(promo.sel?.all),
    nodes: [...(promo.sel?.nodes ?? [])],
    products: [...(promo.sel?.products ?? [])],
  },
  startAt: toDatetimeLocalValue(promo.startAt),
  type: promo.type,
});

export const AdminPromotionsPage = (): JSX.Element => {
  const [
    promotions,
    categories,
    products,
    error,
    isPending,
    isCatalogPending,
    isSaving,
    open,
    save,
    endPromo,
    removePromo,
  ] = useUnit([
    $promotions,
    $promoCategories,
    $promoProducts,
    $promotionsError,
    $isPromotionsPending,
    $isPromoCatalogPending,
    $isPromotionSaving,
    adminPromotionsOpened,
    promotionSaveRequested,
    promotionEndRequested,
    promotionDeleteRequested,
  ]);

  const [editor, setEditor] = useState<EditorDraft | null>(null);
  const [formError, setFormError] = useState<null | string>(null);
  const [confirm, setConfirm] = useState<ConfirmState>(null);

  useEffect(() => {
    open();
  }, [open]);

  const roots = useMemo(() => buildPromoTree(categories, products), [categories, products]);

  const activeCount = promotions.filter((promo) => getPromotionStatus(promo) === 'active').length;
  const scheduledCount = promotions.filter(
    (promo) => getPromotionStatus(promo) === 'scheduled',
  ).length;
  const endedCount = promotions.length - activeCount - scheduledCount;

  const patchEditor = (patch: Partial<EditorDraft>): void => {
    setEditor((previous) => (previous ? { ...previous, ...patch } : previous));
  };

  const submitEditor = (): void => {
    if (!editor) {
      return;
    }

    const name = editor.name.trim();

    if (!name) {
      setFormError('Укажите название акции');

      return;
    }

    if (!editor.startAt || !editor.endAt) {
      setFormError('Укажите период действия');

      return;
    }

    if (new Date(editor.endAt).getTime() <= new Date(editor.startAt).getTime()) {
      setFormError('Окончание должно быть позже начала');

      return;
    }

    const discValue = Number(editor.discValue);

    if (!Number.isFinite(discValue) || discValue <= 0) {
      setFormError('Укажите размер скидки или цену');

      return;
    }

    if (!editor.sel.all && editor.sel.nodes.length === 0 && editor.sel.products.length === 0) {
      setFormError('Выберите товары акции');

      return;
    }

    const payload: PromotionWritePayload = {
      condition: editor.condition.trim(),
      desc: editor.desc.trim(),
      discMode: editor.discMode,
      discValue,
      endAt: editor.endAt,
      minQty: editor.type === 'qty' ? Math.max(1, Number(editor.minQty) || 1) : 0,
      name,
      sel: editor.sel,
      startAt: editor.startAt,
      type: editor.type,
    };

    setFormError(null);
    save({ id: editor.id, payload });
    setEditor(null);
  };

  return (
    <>
      <AdminPageHeader
        actions={
          <Button onPress={() => setEditor(emptyDraft())} variant="primary">
            Создать акцию
          </Button>
        }
        subtitle="Скидки на товары по дате или от количества"
        title="Акции"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <div className={clsx(styles.kpiGrid)}>
        <div className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Всего акций</p>
          <p className={clsx(styles.kpiValue)}>{promotions.length}</p>
        </div>
        <div className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Идут сейчас</p>
          <p className={clsx(styles.kpiValue, promoStyles.kpiActive)}>{activeCount}</p>
        </div>
        <div className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Запланировано</p>
          <p className={clsx(styles.kpiValue, promoStyles.kpiScheduled)}>{scheduledCount}</p>
        </div>
        <div className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Завершены</p>
          <p className={clsx(styles.kpiValue, promoStyles.kpiEnded)}>{endedCount}</p>
        </div>
      </div>

      <div className={clsx(styles.tableWrap)}>
        <table className={clsx(styles.table)}>
          <thead>
            <tr>
              <th>Акция</th>
              <th>Тип</th>
              <th>Скидка</th>
              <th>Период</th>
              <th>Товары</th>
              <th>Статус</th>
              <th aria-label="Действия" />
            </tr>
          </thead>
          <tbody>
            {isPending && promotions.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={7}>
                  Загружаем акции…
                </td>
              </tr>
            ) : null}

            {!isPending && promotions.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={7}>
                  Пока нет акций. Создайте первую акцию по дате или от количества.
                </td>
              </tr>
            ) : null}

            {promotions.map((promo) => {
              const status = getPromotionStatus(promo);

              return (
                <tr key={promo.id}>
                  <td>
                    <strong>{promo.name}</strong>
                    {promo.desc ? <div className={clsx(styles.hint)}>{promo.desc}</div> : null}
                  </td>
                  <td className={clsx(promoStyles.nowrap)}>
                    {PROMOTION_TYPE_LABELS[promo.type]}
                    {promo.type === 'qty' ? (
                      <span className={clsx(styles.hint)}> (от {promo.minQty} шт)</span>
                    ) : null}
                  </td>
                  <td className={clsx(promoStyles.nowrap)}>
                    <strong>{formatPromotionDiscount(promo)}</strong>
                  </td>
                  <td className={clsx(promoStyles.nowrap)}>
                    {formatAdminDateTime(promo.startAt)}
                    <br />
                    {formatAdminDateTime(promo.endAt)}
                  </td>
                  <td>
                    {isCatalogPending ? '…' : summarizePromotionSelection(promo, roots, products)}
                  </td>
                  <td>
                    <span
                      className={clsx(
                        styles.badge,
                        status === 'active' && styles.badgeSuccess,
                        status === 'scheduled' && styles.badgeWarning,
                      )}
                    >
                      {PROMOTION_STATUS_LABELS[status]}
                    </span>
                  </td>
                  <td>
                    <div className={clsx(styles.rowActions)}>
                      <button
                        className={clsx(styles.smallButton)}
                        onClick={() => setEditor(draftFromPromotion(promo))}
                        type="button"
                      >
                        Изменить
                      </button>
                      {status === 'active' ? (
                        <button
                          className={clsx(styles.smallButton, styles.smallButtonDanger)}
                          onClick={() => setConfirm({ kind: 'end', promo })}
                          type="button"
                        >
                          Завершить
                        </button>
                      ) : null}
                      <button
                        className={clsx(styles.smallButton, styles.smallButtonDanger)}
                        onClick={() => setConfirm({ kind: 'delete', promo })}
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

      {editor ? (
        <div
          className={clsx(promoStyles.modalBackdrop)}
          onClick={(event) => {
            if (event.target === event.currentTarget) {
              setEditor(null);
              setFormError(null);
            }
          }}
          role="presentation"
        >
          <div aria-modal="true" className={clsx(promoStyles.modal)} role="dialog">
            <h2 className={clsx(promoStyles.modalTitle)}>
              {editor.id ? 'Редактирование акции' : 'Новая акция'}
            </h2>

            <div className={clsx(styles.form)}>
              <div className={clsx(styles.formGrid)}>
                <label className={clsx(styles.field)}>
                  <span className={clsx(styles.label)}>Название *</span>
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchEditor({ name: event.target.value })}
                    placeholder="Напр. Скидка на метчики"
                    value={editor.name}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  <span className={clsx(styles.label)}>Тип условия *</span>
                  <select
                    className={clsx(styles.input)}
                    onChange={(event) => patchEditor({ type: event.target.value as PromotionType })}
                    value={editor.type}
                  >
                    <option value="date">Без дополнительных условий</option>
                    <option value="qty">От количества товара</option>
                  </select>
                </label>
              </div>

              <label className={clsx(styles.field)}>
                <span className={clsx(styles.label)}>Описание</span>
                <textarea
                  className={clsx(styles.textarea)}
                  onChange={(event) => patchEditor({ desc: event.target.value })}
                  placeholder="Краткое описание для клиента"
                  rows={2}
                  value={editor.desc}
                />
              </label>

              <div className={clsx(styles.formGrid)}>
                <label className={clsx(styles.field)}>
                  <span className={clsx(styles.label)}>Начало *</span>
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchEditor({ startAt: event.target.value })}
                    type="datetime-local"
                    value={editor.startAt}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  <span className={clsx(styles.label)}>Окончание *</span>
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchEditor({ endAt: event.target.value })}
                    type="datetime-local"
                    value={editor.endAt}
                  />
                </label>
              </div>

              <div className={clsx(styles.formGrid)}>
                <label className={clsx(styles.field)}>
                  <span className={clsx(styles.label)}>Тип скидки *</span>
                  <select
                    className={clsx(styles.input)}
                    onChange={(event) =>
                      patchEditor({ discMode: event.target.value as PromotionDiscMode })
                    }
                    value={editor.discMode}
                  >
                    <option value="percent">Скидка, %</option>
                    <option value="price">Акционная цена, ₽</option>
                  </select>
                </label>
                <label className={clsx(styles.field)}>
                  <span className={clsx(styles.label)}>Значение *</span>
                  <input
                    className={clsx(styles.input)}
                    min={0}
                    onChange={(event) => patchEditor({ discValue: event.target.value })}
                    type="number"
                    value={editor.discValue}
                  />
                </label>
                {editor.type === 'qty' ? (
                  <label className={clsx(styles.field)}>
                    <span className={clsx(styles.label)}>Мин. количество позиции *</span>
                    <input
                      className={clsx(styles.input)}
                      min={1}
                      onChange={(event) => patchEditor({ minQty: event.target.value })}
                      type="number"
                      value={editor.minQty}
                    />
                  </label>
                ) : null}
              </div>

              <label className={clsx(styles.field)}>
                <span className={clsx(styles.label)}>Дополнительный текст для клиента</span>
                <input
                  className={clsx(styles.input)}
                  onChange={(event) => patchEditor({ condition: event.target.value })}
                  placeholder="Доп. пояснение к акции"
                  value={editor.condition}
                />
                <span className={clsx(styles.hint)}>
                  Основное условие клиенту сформируется автоматически по типу, скидке и количеству.
                </span>
              </label>

              <PromoProductTree
                onChange={(sel) => patchEditor({ sel })}
                roots={roots}
                selection={editor.sel}
              />

              {formError ? <p className={clsx(styles.error)}>{formError}</p> : null}
            </div>

            <div className={clsx(promoStyles.modalFooter)}>
              <Button
                onPress={() => {
                  setEditor(null);
                  setFormError(null);
                }}
                variant="secondary"
              >
                Отмена
              </Button>
              <Button isDisabled={isSaving} onPress={submitEditor} variant="primary">
                {isSaving ? 'Сохраняем…' : 'Сохранить'}
              </Button>
            </div>
          </div>
        </div>
      ) : null}

      {confirm ? (
        <div
          className={clsx(promoStyles.modalBackdrop)}
          onClick={(event) => {
            if (event.target === event.currentTarget) {
              setConfirm(null);
            }
          }}
          role="presentation"
        >
          <div
            aria-modal="true"
            className={clsx(promoStyles.modal, promoStyles.confirmBox)}
            role="dialog"
          >
            <h2 className={clsx(promoStyles.modalTitle)}>
              {confirm.kind === 'end' ? 'Завершить акцию?' : 'Удалить акцию?'}
            </h2>

            {confirm.kind === 'end' || getPromotionStatus(confirm.promo) === 'active' ? (
              <p className={clsx(promoStyles.warnBanner)}>
                {confirm.kind === 'end'
                  ? `Акция «${confirm.promo.name}» сейчас активна. После завершения скидка перестанет применяться.`
                  : `Акция «${confirm.promo.name}» активна. После удаления скидка больше не применяется.`}
              </p>
            ) : null}

            {confirm.kind === 'delete' && getPromotionStatus(confirm.promo) !== 'active' ? (
              <p className={clsx(styles.hint)}>
                Удалить акцию «{confirm.promo.name}»? Действие необратимо.
              </p>
            ) : null}

            {confirm.kind === 'delete' && getPromotionStatus(confirm.promo) === 'active' ? (
              <p className={clsx(styles.hint)}>
                Удалить акцию «{confirm.promo.name}»? Действие необратимо.
              </p>
            ) : null}

            <div className={clsx(promoStyles.modalFooter)}>
              <Button onPress={() => setConfirm(null)} variant="secondary">
                Отмена
              </Button>
              <Button
                onPress={() => {
                  if (confirm.kind === 'end') {
                    endPromo(confirm.promo.id);
                  } else {
                    removePromo(confirm.promo.id);
                  }

                  setConfirm(null);
                }}
                variant="danger"
              >
                {confirm.kind === 'end' ? 'Завершить сейчас' : 'Удалить'}
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
};
