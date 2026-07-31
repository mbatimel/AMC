'use client';

import { Button, FieldError, Form, Input, Label, TextArea, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useSearchParams } from 'next/navigation';
import { Suspense, useEffect, useState } from 'react';

import { useSession } from '@/core/entities/session';
import { SUPPORT_SEVERITY_OPTIONS, SUPPORT_STATUS_LABELS } from '@/core/shared/api/support';
import { readFormString } from '@/core/shared/lib/readFormString';
import { FormSelect } from '@/core/shared/ui/FormSelect';
import { InfoCard, InfoPage } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

import styles from './Support.module.css';
import {
  $isSupportPending,
  $isSupportSent,
  $mySupportRequests,
  $supportError,
  $supportOrders,
  supportFormReset,
  supportOpened,
  supportSubmitted,
} from './model';

const NO_ORDER = 'none';

const formatDateTime = (value: string): string => {
  const date = new Date(value);

  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat('ru-RU', {
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        month: '2-digit',
        year: 'numeric',
      }).format(date);
};

const SupportContent = (): JSX.Element => {
  const searchParams = useSearchParams();
  const { userId } = useSession();
  const prefillOrder = searchParams.get('order') ?? '';
  const [severity, setSeverity] = useState('3');
  const [orderId, setOrderId] = useState(prefillOrder || NO_ORDER);

  const [orders, requests, isPending, isSent, error, open, submit, reset] = useUnit([
    $supportOrders,
    $mySupportRequests,
    $isSupportPending,
    $isSupportSent,
    $supportError,
    supportOpened,
    supportSubmitted,
    supportFormReset,
  ]);

  useEffect(() => {
    open();
  }, [open]);

  useEffect(() => {
    if (prefillOrder) {
      setOrderId(prefillOrder);
    }
  }, [prefillOrder]);

  const orderOptions = [
    { label: 'Не связан с заказом', value: NO_ORDER },
    ...orders.map((order) => ({
      label: `${order.number || order.id} · ${formatDateTime(order.created_at)}`,
      value: order.id,
    })),
  ];

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const form = event.currentTarget;
    const formData = new FormData(form);
    const subject = readFormString(formData, 'subject').trim();
    const text = readFormString(formData, 'text').trim();
    const contact = readFormString(formData, 'contact').trim();

    if (!subject || !text) {
      return;
    }

    submit({
      contact,
      orderId: orderId === NO_ORDER ? '' : orderId,
      severity: Number(severity) || 3,
      source: 'form',
      subject,
      text,
      userId: userId ?? undefined,
    });

    form.reset();
  };

  return (
    <InfoPage
      description="Опишите вопрос или проблему — обращение попадёт менеджеру и в журнал портала."
      eyebrow="Поддержка"
      title="Поддержка"
    >
      <InfoCard title="Написать в поддержку">
        {isSent ? (
          <div className={clsx(styles.success)}>
            <p className={clsx(styles.successTitle)}>Обращение зарегистрировано</p>
            <p className={clsx(styles.successText)}>
              Менеджер свяжется с вами по указанным контактам. Статус обращения виден в списке ниже.
            </p>
            <Button onPress={() => reset()} variant="outline">
              Отправить ещё одно
            </Button>
          </div>
        ) : (
          <Form className={clsx(styles.form)} onSubmit={handleSubmit}>
            <TextField className={clsx(styles.field)} isRequired name="subject">
              <Label className={clsx(styles.label)}>Название проблемы</Label>
              <Input
                className={clsx(styles.input)}
                fullWidth
                placeholder="Например: не открывается документ по заказу"
              />
              <FieldError />
            </TextField>

            <TextField className={clsx(styles.field)} isRequired name="text">
              <Label className={clsx(styles.label)}>Описание проблемы</Label>
              <TextArea
                className={clsx(styles.textarea)}
                fullWidth
                placeholder="Опишите, что произошло и какой результат ожидаете"
              />
              <FieldError />
            </TextField>

            <div className={clsx(styles.row)}>
              <FormSelect
                ariaLabel="Срочность обращения"
                label="Срочность"
                onChange={setSeverity}
                options={SUPPORT_SEVERITY_OPTIONS.map((option) => ({
                  label: option.label,
                  value: String(option.value),
                }))}
                value={severity}
              />
              <FormSelect
                ariaLabel="Связанный заказ"
                label="Связанный заказ"
                onChange={setOrderId}
                options={orderOptions}
                value={orderId}
              />
            </div>

            <TextField className={clsx(styles.field)} name="contact">
              <Label className={clsx(styles.label)}>Контакт для ответа</Label>
              <Input className={clsx(styles.input)} fullWidth placeholder="E-mail или телефон" />
            </TextField>

            {error ? <p className={clsx(styles.error)}>{error}</p> : null}

            <Button isDisabled={isPending} type="submit" variant="primary">
              {isPending ? 'Отправляем…' : 'Отправить обращение'}
            </Button>
          </Form>
        )}
      </InfoCard>

      {requests.length > 0 ? (
        <InfoCard title="Мои обращения">
          <ul className={clsx(styles.requests)}>
            {requests.map((request) => (
              <li className={clsx(styles.request)} key={request.id}>
                <div className={clsx(styles.requestHead)}>
                  <span className={clsx(styles.requestSubject)}>{request.subject}</span>
                  <span className={clsx(styles.requestStatus)}>
                    {SUPPORT_STATUS_LABELS[request.status]}
                  </span>
                </div>
                <p className={clsx(styles.requestMeta)}>
                  {formatDateTime(request.created_at)}
                  {request.order_id ? ` · заказ ${request.order_id}` : ''}
                  {` · срочность ${request.severity}`}
                </p>
                <p className={clsx(styles.requestText)}>{request.text}</p>
                {request.answer ? (
                  <p className={clsx(styles.requestAnswer)}>Ответ: {request.answer}</p>
                ) : null}
              </li>
            ))}
          </ul>
        </InfoCard>
      ) : null}
    </InfoPage>
  );
};

export const Support = (): JSX.Element => {
  return (
    <Page>
      <Suspense fallback={<></>}>
        <SupportContent />
      </Suspense>
    </Page>
  );
};
