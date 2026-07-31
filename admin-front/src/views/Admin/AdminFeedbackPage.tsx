'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { ORDER_RATING_MAX } from '@/core/shared/api/feedback';
import { FormSelect } from '@/core/shared/ui/FormSelect';

import styles from './Admin.module.css';
import { formatAdminDateTime } from './lib/nav';
import {
  $adminFeedback,
  $adminFeedbackError,
  $feedbackFilters,
  adminFeedbackOpened,
  feedbackFiltersChanged,
} from './model/feedback';
import { AdminPageHeader } from './ui/AdminPageHeader';

const ANY_RATING = 'any';

const RATING_OPTIONS = [
  { label: 'Все оценки', value: ANY_RATING },
  ...[5, 4, 3, 2, 1].map((rating) => ({ label: `${rating} из 5`, value: String(rating) })),
];

const toDateValue = (value: string): string => value.slice(0, 10);

export const AdminFeedbackPage = (): JSX.Element => {
  const [feedback, filters, error, open, changeFilters] = useUnit([
    $adminFeedback,
    $feedbackFilters,
    $adminFeedbackError,
    adminFeedbackOpened,
    feedbackFiltersChanged,
  ]);

  useEffect(() => {
    open();
  }, [open]);

  const rows = feedback.filter((item) => {
    const date = toDateValue(item.created_at);

    if (filters.from && date < filters.from) {
      return false;
    }

    if (filters.to && date > filters.to) {
      return false;
    }

    if (filters.rating && String(item.rating) !== filters.rating) {
      return false;
    }

    if (filters.query) {
      const query = filters.query.toLowerCase();

      return (
        item.order_id.toLowerCase().includes(query) ||
        item.client_name.toLowerCase().includes(query) ||
        item.text.toLowerCase().includes(query)
      );
    }

    return true;
  });

  const average =
    feedback.length > 0
      ? (feedback.reduce((sum, item) => sum + item.rating, 0) / feedback.length).toFixed(1)
      : '—';
  const fiveStars = feedback.filter((item) => item.rating === ORDER_RATING_MAX).length;

  return (
    <>
      <AdminPageHeader
        subtitle={`${rows.length} из ${feedback.length} отзывов`}
        title="Отзывы по заказам"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <div className={clsx(styles.kpiGrid)}>
        <article className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Всего отзывов</p>
          <p className={clsx(styles.kpiValue)}>{feedback.length}</p>
        </article>
        <article className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Средняя оценка</p>
          <p className={clsx(styles.kpiValue)}>{average}</p>
          <p className={clsx(styles.kpiHint)}>из {ORDER_RATING_MAX}</p>
        </article>
        <article className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Оценка 5/5</p>
          <p className={clsx(styles.kpiValue)}>{fiveStars}</p>
        </article>
      </div>

      <section className={clsx(styles.card)}>
        <div className={clsx(styles.filters)}>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="feedback-from">
              Дата с
            </label>
            <input
              className={clsx(styles.input)}
              id="feedback-from"
              onChange={(event) => changeFilters({ from: event.target.value })}
              type="date"
              value={filters.from}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="feedback-to">
              Дата по
            </label>
            <input
              className={clsx(styles.input)}
              id="feedback-to"
              onChange={(event) => changeFilters({ to: event.target.value })}
              type="date"
              value={filters.to}
            />
          </div>
          <FormSelect
            ariaLabel="Фильтр по оценке"
            label="Оценка"
            onChange={(rating) => changeFilters({ rating: rating === ANY_RATING ? '' : rating })}
            options={RATING_OPTIONS}
            value={filters.rating || ANY_RATING}
          />
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="feedback-query">
              Клиент, заказ или текст
            </label>
            <input
              className={clsx(styles.input)}
              id="feedback-query"
              onChange={(event) => changeFilters({ query: event.target.value })}
              placeholder="ORD-2026 или название компании"
              value={filters.query}
            />
          </div>
        </div>
      </section>

      <div className={clsx(styles.tableWrap)}>
        <table className={clsx(styles.table)}>
          <thead>
            <tr>
              <th>Дата</th>
              <th>Заказ</th>
              <th>Клиент</th>
              <th>Оценка</th>
              <th>Отзыв</th>
            </tr>
          </thead>
          <tbody>
            {rows.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={5}>
                  Отзывы не найдены
                </td>
              </tr>
            ) : null}

            {rows.map((item) => (
              <tr key={item.id}>
                <td>{formatAdminDateTime(item.created_at)}</td>
                <td>{item.order_id}</td>
                <td>{item.client_name || item.user_id || '—'}</td>
                <td>
                  {'★'.repeat(item.rating)}
                  {'☆'.repeat(Math.max(0, ORDER_RATING_MAX - item.rating))}
                </td>
                <td>{item.text || 'Без комментария'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
};
