'use client';

import { Alert, Button, Description, EmptyState, Spinner, Typography } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useRouter } from 'next/navigation';

import { IconPlus } from '@/core/shared/icons';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { formatOrdersCount } from '@/core/shared/lib/pluralize';
import { AppPath, getCabinetOrderPath } from '@/core/shared/router/paths';

import styles from '../Cabinet.module.css';
import {
  formatDeliveryType,
  formatOrderDate,
  formatOrderStatus,
  formatPaymentStatus,
  paymentTone,
  statusTone,
} from '../lib/labels';
import {
  $hasMoreOrders,
  $isOrdersPending,
  $orders,
  $ordersError,
  $ordersTotal,
  cabinetOrdersLoadMore,
} from '../model/orders';

const statusClassName = (status: string): string => {
  const map = {
    cancelled: styles.statusCancelled,
    completed: styles.statusCompleted,
    default: styles.statusDefault,
    new: styles.statusNew,
    processing: styles.statusProcessing,
    shipped: styles.statusShipped,
  } as const;

  return map[statusTone(status)];
};

const paymentClassName = (status: string): string => {
  const map = {
    default: styles.paymentDefault,
    paid: styles.paymentPaid,
    pending: styles.paymentPending,
    unpaid: styles.paymentUnpaid,
  } as const;

  return map[paymentTone(status)];
};

export const CabinetOrders = (): JSX.Element => {
  const router = useRouter();
  const [orders, total, pending, error, hasMore, loadMore] = useUnit([
    $orders,
    $ordersTotal,
    $isOrdersPending,
    $ordersError,
    $hasMoreOrders,
    cabinetOrdersLoadMore,
  ]);

  return (
    <div className={clsx(styles.main)}>
      <div className={clsx(styles.pageHeader)}>
        <div>
          <Typography.Heading className={clsx(styles.pageTitle)} level={1}>
            Мои заказы
          </Typography.Heading>
          <Description className={clsx(styles.pageSubtitle)}>
            {formatOrdersCount(total)}
          </Description>
        </div>
        <Button
          className={clsx(styles.newOrderButton)}
          onPress={() => router.push(AppPath.Catalog)}
          variant="primary"
        >
          <IconPlus currentColor="currentColor" height={16} width={16} />
          Новый заказ
        </Button>
      </div>

      {error ? (
        <Alert className={clsx(styles.alert)} status="danger">
          <Alert.Indicator />
          <Alert.Content>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      ) : null}

      {!pending && orders.length === 0 && !error ? (
        <EmptyState className={clsx(styles.empty)}>
          <Description>Пока нет заказов</Description>
        </EmptyState>
      ) : null}

      {orders.length > 0 ? (
        <div className={clsx(styles.ordersScroll)}>
          <table className={clsx(styles.ordersTable)}>
            <thead>
              <tr>
                <th>Номер</th>
                <th>Дата</th>
                <th>Позиций</th>
                <th>Сумма</th>
                <th>Доставка</th>
                <th>Статус</th>
                <th>Оплата</th>
                <th>Документы</th>
                <th aria-label="Действия" />
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => (
                <tr key={order.id}>
                  <td>
                    <span className={clsx(styles.orderNumber)}>
                      {order.number || order.id.slice(0, 8)}
                    </span>
                  </td>
                  <td>
                    <span className={clsx(styles.orderMeta)}>
                      {formatOrderDate(order.created_at)}
                    </span>
                  </td>
                  <td>
                    <span className={clsx(styles.orderMeta)}>{order.items.length}</span>
                  </td>
                  <td>
                    <span className={clsx(styles.orderTotal)}>{formatPrice(order.total)}</span>
                  </td>
                  <td>
                    <span className={clsx(styles.orderMeta)}>
                      {formatDeliveryType(order.delivery_type)}
                    </span>
                  </td>
                  <td>
                    <span className={clsx(styles.statusBadge, statusClassName(order.status))}>
                      {formatOrderStatus(order.status)}
                    </span>
                  </td>
                  <td>
                    <span
                      className={clsx(styles.statusBadge, paymentClassName(order.payment_status))}
                    >
                      {formatPaymentStatus(order.payment_status)}
                    </span>
                  </td>
                  <td>
                    <span className={clsx(styles.orderMeta)}>{order.documents.length}</span>
                  </td>
                  <td>
                    <Button
                      className={clsx(styles.openButton)}
                      onPress={() => router.push(getCabinetOrderPath(order.id))}
                      size="sm"
                      variant="outline"
                    >
                      Открыть
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <ul className={clsx(styles.ordersCards)}>
            {orders.map((order) => (
              <li className={clsx(styles.orderCard)} key={order.id}>
                <div className={clsx(styles.orderCardHeader)}>
                  <div>
                    <p className={clsx(styles.orderNumber)}>
                      {order.number || order.id.slice(0, 8)}
                    </p>
                    <p className={clsx(styles.orderMeta)}>{formatOrderDate(order.created_at)}</p>
                  </div>
                  <p className={clsx(styles.orderTotal)}>{formatPrice(order.total)}</p>
                </div>

                <div className={clsx(styles.orderCardBadges)}>
                  <span className={clsx(styles.statusBadge, statusClassName(order.status))}>
                    {formatOrderStatus(order.status)}
                  </span>
                  <span
                    className={clsx(styles.statusBadge, paymentClassName(order.payment_status))}
                  >
                    {formatPaymentStatus(order.payment_status)}
                  </span>
                </div>

                <p className={clsx(styles.orderCardMeta)}>
                  {order.items.length} поз. · {formatDeliveryType(order.delivery_type)}
                  {order.documents.length > 0 ? ` · док.: ${order.documents.length}` : ''}
                </p>

                <Button
                  className={clsx(styles.openButton, styles.orderCardOpen)}
                  onPress={() => router.push(getCabinetOrderPath(order.id))}
                  size="sm"
                  variant="outline"
                >
                  Открыть
                </Button>
              </li>
            ))}
          </ul>

          {hasMore ? (
            <div className={clsx(styles.loadMore)}>
              <Button isDisabled={pending} onPress={() => loadMore()} variant="secondary">
                Показать ещё
              </Button>
            </div>
          ) : null}
        </div>
      ) : null}

      {pending ? (
        <div className={clsx(styles.statusRow)}>
          <Spinner size="sm" />
          <Description>Загрузка…</Description>
        </div>
      ) : null}
    </div>
  );
};
