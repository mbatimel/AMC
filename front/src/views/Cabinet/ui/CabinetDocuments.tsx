'use client';

import { Alert, Button, Description, EmptyState, Spinner, Typography } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useRouter } from 'next/navigation';

import { AppPath } from '@/core/shared/router/paths';

import styles from '../Cabinet.module.css';
import extraStyles from '../CabinetExtras.module.css';
import { formatOrderDate } from '../lib/labels';
import { $cabinetDocuments, formatDocumentType } from '../model/documents';
import { $isOrdersPending, $ordersError } from '../model/orders';

export const CabinetDocuments = (): JSX.Element => {
  const router = useRouter();
  const [documents, isPending, error] = useUnit([
    $cabinetDocuments,
    $isOrdersPending,
    $ordersError,
  ]);

  return (
    <div className={clsx(styles.main)}>
      <div className={clsx(styles.pageHeader)}>
        <div>
          <Typography.Heading className={clsx(styles.pageTitle)} level={1}>
            Документы
          </Typography.Heading>
          <Description className={clsx(styles.pageSubtitle)}>
            Счета, накладные и УПД по всем заказам
          </Description>
        </div>
      </div>

      {error ? (
        <Alert className={clsx(styles.alert)} status="danger">
          <Alert.Indicator />
          <Alert.Content>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      ) : null}

      {isPending && documents.length === 0 ? <Spinner /> : null}

      {!isPending && documents.length === 0 ? (
        <EmptyState className={clsx(extraStyles.emptyState)}>
          <Description>
            Документов пока нет. Счёт появится после подтверждения заказа менеджером, закрывающие
            документы — после отгрузки.
          </Description>
          <Button onPress={() => router.push(AppPath.CabinetOrders)} variant="primary">
            Перейти к заказам
          </Button>
        </EmptyState>
      ) : null}

      {documents.length > 0 ? (
        <div className={clsx(extraStyles.tableWrap)}>
          <table className={clsx(extraStyles.table)}>
            <thead>
              <tr>
                <th>Дата</th>
                <th>Тип</th>
                <th>Номер</th>
                <th>Заказ</th>
                <th aria-label="Действия" />
              </tr>
            </thead>
            <tbody>
              {documents.map((document) => (
                <tr key={document.id}>
                  <td>{formatOrderDate(document.created_at)}</td>
                  <td>{formatDocumentType(document.type)}</td>
                  <td>{document.name || '—'}</td>
                  <td>{document.orderNumber}</td>
                  <td>
                    {document.url ? (
                      <a
                        className={clsx(extraStyles.link)}
                        href={document.url}
                        rel="noreferrer"
                        target="_blank"
                      >
                        Скачать
                      </a>
                    ) : (
                      <span className={clsx(extraStyles.muted)}>Готовится</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </div>
  );
};
