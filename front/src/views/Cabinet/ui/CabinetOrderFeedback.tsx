'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect, useState } from 'react';

import { useSession } from '@/core/entities/session';
import { ORDER_RATING_MAX } from '@/core/shared/api/feedback';
import { getSupportPath } from '@/core/shared/router/paths';

import styles from '../CabinetExtras.module.css';
import {
  $feedbackError,
  $isFeedbackPending,
  $orderFeedback,
  orderFeedbackRequested,
  orderFeedbackSubmitted,
} from '../model/feedback';

type CabinetOrderFeedbackProps = {
  orderId: string;
  orderNumber: string;
  orderStatus: string;
};

const RATING_VALUES = Array.from({ length: ORDER_RATING_MAX }, (_, index) => index + 1);

/** Оценка завершённого заказа + быстрый переход в поддержку по этому заказу. */
export const CabinetOrderFeedback = ({
  orderId,
  orderNumber,
  orderStatus,
}: CabinetOrderFeedbackProps): JSX.Element | null => {
  const { userId } = useSession();
  const [feedback, isPending, error, request, submit] = useUnit([
    $orderFeedback,
    $isFeedbackPending,
    $feedbackError,
    orderFeedbackRequested,
    orderFeedbackSubmitted,
  ]);
  const [rating, setRating] = useState(0);
  const [text, setText] = useState('');

  useEffect(() => {
    request();
  }, [request]);

  const existing = feedback.find((item) => item.order_id === orderId);
  const isCompleted = orderStatus === 'completed' || orderStatus === 'delivered';

  if (!isCompleted && !existing) {
    return (
      <section className={clsx(styles.feedbackCard)}>
        <h2 className={clsx(styles.feedbackTitle)}>Нужна помощь по заказу?</h2>
        <p className={clsx(styles.feedbackHint)}>
          Задайте вопрос менеджеру — обращение будет привязано к заказу {orderNumber}.
        </p>
        <Link className={clsx(styles.link)} href={getSupportPath(orderId)}>
          Написать в поддержку
        </Link>
      </section>
    );
  }

  return (
    <section className={clsx(styles.feedbackCard)}>
      <h2 className={clsx(styles.feedbackTitle)}>Оценка заказа</h2>

      {existing ? (
        <>
          <p className={clsx(styles.feedbackStars)} aria-label={`Оценка ${existing.rating} из 5`}>
            {'★'.repeat(existing.rating)}
            {'☆'.repeat(Math.max(0, ORDER_RATING_MAX - existing.rating))}
          </p>
          <p className={clsx(styles.feedbackText)}>{existing.text || 'Без комментария'}</p>
        </>
      ) : (
        <>
          <p className={clsx(styles.feedbackHint)}>
            Заказ завершён — расскажите, как всё прошло.
          </p>
          <div className={clsx(styles.ratingRow)} role="radiogroup" aria-label="Оценка заказа">
            {RATING_VALUES.map((value) => (
              <button
                aria-checked={rating === value}
                aria-label={`${value} из 5`}
                className={clsx(styles.ratingButton, rating >= value && styles.ratingButtonActive)}
                key={value}
                onClick={() => setRating(value)}
                role="radio"
                type="button"
              >
                ★
              </button>
            ))}
          </div>
          <textarea
            aria-label="Комментарий к заказу"
            className={clsx(styles.feedbackTextarea)}
            onChange={(event) => setText(event.target.value)}
            placeholder="Что понравилось, а что можно улучшить"
            rows={3}
            value={text}
          />
          {error ? <p className={clsx(styles.feedbackError)}>{error}</p> : null}
          <Button
            isDisabled={rating === 0 || isPending}
            onPress={() =>
              submit({
                orderId,
                orderStatus,
                rating,
                text,
                userId: userId ?? undefined,
              })
            }
            variant="primary"
          >
            {isPending ? 'Отправляем…' : 'Отправить отзыв'}
          </Button>
        </>
      )}

      <Link className={clsx(styles.link)} href={getSupportPath(orderId)}>
        Написать в поддержку по заказу
      </Link>
    </section>
  );
};
