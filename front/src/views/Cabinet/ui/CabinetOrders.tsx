'use client';

import {
  Alert,
  Button,
  Chip,
  Description,
  EmptyState,
  Link as HeroLink,
  Spinner,
  Surface,
  Table,
  Typography,
} from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useRouter } from 'next/navigation';

import type { Order } from '@/core/shared/api/orders';

import { IconPlus } from '@/core/shared/icons';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { formatOrdersCount } from '@/core/shared/lib/pluralize';
import { AppPath } from '@/core/shared/router/paths';

import styles from '../Cabinet.module.css';
import {
  formatDeliveryType,
  formatOrderDate,
  formatOrderStatus,
  formatPaymentStatus,
  paymentChipColor,
  statusChipColor,
} from '../lib/labels';
import {
  $hasMoreOrders,
  $isOrdersPending,
  $orders,
  $ordersError,
  $ordersTotal,
  $selectedOrder,
  $selectedOrderId,
  cabinetOrderSelected,
  cabinetOrdersLoadMore,
} from '../model/orders';

type OrderDetailsProps = {
  order: Order;
};

const OrderDetails = ({ order }: OrderDetailsProps): JSX.Element => {
  return (
    <div className={clsx(styles.orderDetails)}>
      <dl className={clsx(styles.detailsGrid)}>
        <dt>Доставка</dt>
        <dd>{formatDeliveryType(order.delivery_type)}</dd>
        {order.delivery_address ? (
          <>
            <dt>Адрес</dt>
            <dd>{order.delivery_address}</dd>
          </>
        ) : null}
        {order.contact_name ? (
          <>
            <dt>Контакт</dt>
            <dd>{order.contact_name}</dd>
          </>
        ) : null}
        {order.phone ? (
          <>
            <dt>Телефон</dt>
            <dd>{order.phone}</dd>
          </>
        ) : null}
        {order.comment ? (
          <>
            <dt>Комментарий</dt>
            <dd>{order.comment}</dd>
          </>
        ) : null}
      </dl>

      {order.items.length > 0 ? (
        <Table className={clsx(styles.itemsTable)}>
          <Table.ScrollContainer>
            <Table.Content aria-label="Состав заказа">
              <Table.Header>
                <Table.Column isRowHeader>Товар</Table.Column>
                <Table.Column>Артикул</Table.Column>
                <Table.Column>Кол-во</Table.Column>
                <Table.Column>Цена</Table.Column>
                <Table.Column>Сумма</Table.Column>
              </Table.Header>
              <Table.Body items={order.items}>
                {(item) => (
                  <Table.Row id={item.id}>
                    <Table.Cell>{item.product_name || '—'}</Table.Cell>
                    <Table.Cell>{item.sku || '—'}</Table.Cell>
                    <Table.Cell>{item.qty}</Table.Cell>
                    <Table.Cell>{formatPrice(item.price)}</Table.Cell>
                    <Table.Cell>{formatPrice(item.total)}</Table.Cell>
                  </Table.Row>
                )}
              </Table.Body>
            </Table.Content>
          </Table.ScrollContainer>
        </Table>
      ) : (
        <Description>Состав заказа пуст</Description>
      )}

      {order.documents.length > 0 ? (
        <div className={clsx(styles.docs)}>
          {order.documents.map((doc) =>
            doc.url ? (
              <HeroLink href={doc.url} key={doc.id} rel="noreferrer" target="_blank">
                {doc.name || doc.type || 'Документ'}
              </HeroLink>
            ) : (
              <Description key={doc.id}>{doc.name || doc.type || 'Документ'}</Description>
            ),
          )}
        </div>
      ) : null}
    </div>
  );
};

export const CabinetOrders = (): JSX.Element => {
  const router = useRouter();
  const [orders, selectedId, selectedOrder, total, pending, error, hasMore, loadMore, select] =
    useUnit([
      $orders,
      $selectedOrderId,
      $selectedOrder,
      $ordersTotal,
      $isOrdersPending,
      $ordersError,
      $hasMoreOrders,
      cabinetOrdersLoadMore,
      cabinetOrderSelected,
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
        <Surface className={clsx(styles.panel)}>
          <Table className={clsx(styles.ordersTable)}>
            <Table.ScrollContainer>
              <Table.Content aria-label="Мои заказы">
                <Table.Header>
                  <Table.Column isRowHeader>Номер</Table.Column>
                  <Table.Column>Дата</Table.Column>
                  <Table.Column>Позиций</Table.Column>
                  <Table.Column>Сумма</Table.Column>
                  <Table.Column>Доставка</Table.Column>
                  <Table.Column>Статус</Table.Column>
                  <Table.Column>Оплата</Table.Column>
                  <Table.Column>Документы</Table.Column>
                  <Table.Column> </Table.Column>
                </Table.Header>
                <Table.Body items={orders}>
                  {(order) => (
                    <Table.Row
                      className={clsx(selectedId === order.id && styles.ordersRowActive)}
                      id={order.id}
                    >
                      <Table.Cell>
                        <span className={clsx(styles.orderNumber)}>
                          {order.number || order.id.slice(0, 8)}
                        </span>
                      </Table.Cell>
                      <Table.Cell>
                        <span className={clsx(styles.orderMeta)}>
                          {formatOrderDate(order.created_at)}
                        </span>
                      </Table.Cell>
                      <Table.Cell>{order.items.length}</Table.Cell>
                      <Table.Cell>
                        <span className={clsx(styles.orderTotal)}>{formatPrice(order.total)}</span>
                      </Table.Cell>
                      <Table.Cell>{formatDeliveryType(order.delivery_type)}</Table.Cell>
                      <Table.Cell>
                        <Chip color={statusChipColor(order.status)} size="sm" variant="soft">
                          <Chip.Label>{formatOrderStatus(order.status)}</Chip.Label>
                        </Chip>
                      </Table.Cell>
                      <Table.Cell>
                        <Chip
                          color={paymentChipColor(order.payment_status)}
                          size="sm"
                          variant="soft"
                        >
                          <Chip.Label>{formatPaymentStatus(order.payment_status)}</Chip.Label>
                        </Chip>
                      </Table.Cell>
                      <Table.Cell>{order.documents.length}</Table.Cell>
                      <Table.Cell>
                        <Button
                          className={clsx(styles.openButton)}
                          onPress={() => select(selectedId === order.id ? '' : order.id)}
                          size="sm"
                          variant="outline"
                        >
                          Открыть
                        </Button>
                      </Table.Cell>
                    </Table.Row>
                  )}
                </Table.Body>
              </Table.Content>
            </Table.ScrollContainer>
          </Table>

          {selectedOrder ? (
            <div className={clsx(styles.orderDetailsPanel)}>
              <Typography.Heading className={clsx(styles.sectionTitle)} level={3}>
                Заказ {selectedOrder.number || selectedOrder.id.slice(0, 8)}
              </Typography.Heading>
              <OrderDetails order={selectedOrder} />
            </div>
          ) : null}

          {hasMore ? (
            <div className={clsx(styles.loadMore)}>
              <Button isDisabled={pending} onPress={() => loadMore()} variant="secondary">
                Показать ещё
              </Button>
            </div>
          ) : null}
        </Surface>
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
