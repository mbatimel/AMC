'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect, useState } from 'react';

import { SUPPORT_STATUS_LABELS } from '@/core/shared/api/support';

import styles from './Admin.module.css';
import { formatAdminDateTime } from './lib/nav';
import {
  $adminFeedbackError,
  $adminSupport,
  adminSupportOpened,
  supportRequestUpdated,
} from './model/feedback';
import { AdminPageHeader } from './ui/AdminPageHeader';

export const AdminSupportPage = (): JSX.Element => {
  const [requests, error, open, update] = useUnit([
    $adminSupport,
    $adminFeedbackError,
    adminSupportOpened,
    supportRequestUpdated,
  ]);
  const [answers, setAnswers] = useState<Record<string, string>>({});

  useEffect(() => {
    open();
  }, [open]);

  const openCount = requests.filter((request) => request.status !== 'closed').length;

  return (
    <>
      <AdminPageHeader
        subtitle={`${openCount} открытых из ${requests.length}`}
        title="Обращения поддержки"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      {requests.length === 0 ? (
        <section className={clsx(styles.card)}>
          <p className={clsx(styles.hint)}>
            Обращений пока нет. Они появляются из формы «Поддержка» и при вызове оператора в чате
            подбора инструмента.
          </p>
        </section>
      ) : null}

      {requests.map((request) => (
        <section className={clsx(styles.card)} key={request.id}>
          <div className={clsx(styles.pageHeader)}>
            <div>
              <h2 className={clsx(styles.cardTitle)}>{request.subject}</h2>
              <p className={clsx(styles.hint)}>
                {formatAdminDateTime(request.created_at)} · срочность {request.severity}
                {request.order_id ? ` · заказ ${request.order_id}` : ''}
                {request.source === 'assistant' ? ' · из чата подбора' : ''}
              </p>
            </div>
            <span
              className={clsx(
                styles.badge,
                request.status === 'new' && styles.badgeWarning,
                request.status === 'closed' && styles.badgeSuccess,
              )}
            >
              {SUPPORT_STATUS_LABELS[request.status]}
            </span>
          </div>

          <p className={clsx(styles.hint)} style={{ whiteSpace: 'pre-wrap' }}>
            {request.text}
          </p>

          {request.contact ? (
            <p className={clsx(styles.hint)}>Контакт: {request.contact}</p>
          ) : null}

          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor={`answer-${request.id}`}>
              Ответ клиенту
            </label>
            <textarea
              className={clsx(styles.textarea)}
              id={`answer-${request.id}`}
              onChange={(event) =>
                setAnswers((previous) => ({ ...previous, [request.id]: event.target.value }))
              }
              rows={3}
              value={answers[request.id] ?? request.answer}
            />
          </div>

          <div className={clsx(styles.rowActions)}>
            <button
              className={clsx(styles.smallButton)}
              onClick={() =>
                update({
                  answer: answers[request.id] ?? request.answer,
                  id: request.id,
                  status: 'in_progress',
                })
              }
              type="button"
            >
              Взять в работу
            </button>
            <button
              className={clsx(styles.smallButton, styles.smallButtonPrimary)}
              onClick={() =>
                update({
                  answer: answers[request.id] ?? request.answer,
                  id: request.id,
                  status: 'closed',
                })
              }
              type="button"
            >
              Закрыть обращение
            </button>
          </div>
        </section>
      ))}
    </>
  );
};
