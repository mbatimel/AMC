'use client';

import { Alert, Button, Description, Spinner, Typography } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect } from 'react';

import type { Order, OrderHistoryItem } from '@/core/shared/api/orders';

import { formatPrice } from '@/core/shared/lib/formatPrice';
import { AppPath } from '@/core/shared/router/paths';

import listStyles from '../Cabinet.module.css';
import {
  formatDeliveryType,
  formatOrderDate,
  formatOrderDateTime,
  formatOrderStatus,
  formatPaymentStatus,
  paymentTone,
  statusTone,
} from '../lib/labels';
import {
  $isOrderDetailPending,
  $orderDetail,
  $orderDetailError,
  cabinetOrderClosed,
  cabinetOrderOpened,
} from '../model/orderDetail';
import { $profile } from '../model/profile';
import { CabinetOrderFeedback } from './CabinetOrderFeedback';
import styles from './CabinetOrder.module.css';

type CabinetOrderProps = {
  orderId: string;
};

const statusClassName = (status: string): string => {
  const map = {
    cancelled: listStyles.statusCancelled,
    completed: listStyles.statusCompleted,
    default: listStyles.statusDefault,
    new: listStyles.statusNew,
    processing: listStyles.statusProcessing,
    shipped: listStyles.statusShipped,
  } as const;

  return map[statusTone(status)];
};

const paymentClassName = (status: string): string => {
  const map = {
    default: listStyles.paymentDefault,
    paid: listStyles.paymentPaid,
    pending: listStyles.paymentPending,
    unpaid: listStyles.paymentUnpaid,
  } as const;

  return map[paymentTone(status)];
};

const displayValue = (value: string): string => (value.trim().length > 0 ? value : '—');

const historyLabel = (item: OrderHistoryItem): string => {
  if (item.comment.trim().length > 0) {
    return item.comment;
  }

  const status = formatOrderStatus(item.status);

  if (status !== '—') {
    return status;
  }

  return 'Обновление статуса';
};

const buildHistory = (order: Order): OrderHistoryItem[] => {
  if (order.history.length > 0) {
    return order.history;
  }

  if (!order.created_at) {
    return [];
  }

  return [
    {
      changed_by: '',
      comment: 'Заказ создан',
      created_at: order.created_at,
      id: `${order.id}-created`,
      order_id: order.id,
      payment_status: order.payment_status,
      status: order.status || 'new',
    },
  ];
};

export const CabinetOrder = ({ orderId }: CabinetOrderProps): JSX.Element => {
  const [order, pending, error, profile, open, close] = useUnit([
    $orderDetail,
    $isOrderDetailPending,
    $orderDetailError,
    $profile,
    cabinetOrderOpened,
    cabinetOrderClosed,
  ]);

  useEffect(() => {
    open(orderId);

    return () => {
      close();
    };
  }, [close, open, orderId]);

  const companyName =
    profile?.active_client?.company_name ||
    [profile?.last_name, profile?.first_name].filter(Boolean).join(' ') ||
    '';

  if (pending && !order) {
    return (
      <div className={clsx(styles.root)}>
        <div className={clsx(styles.statusRow)}>
          <Spinner size="sm" />
          <Description>Загрузка заказа…</Description>
        </div>
      </div>
    );
  }

  if (error && !order) {
    return (
      <div className={clsx(styles.root)}>
        <Link className={clsx(styles.backLink)} href={AppPath.CabinetOrders}>
          ← Мои заказы
        </Link>
        <Alert className={clsx(listStyles.alert)} status="danger">
          <Alert.Indicator />
          <Alert.Content>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      </div>
    );
  }

  if (!order) {
    return (
      <div className={clsx(styles.root)}>
        <Link className={clsx(styles.backLink)} href={AppPath.CabinetOrders}>
          ← Мои заказы
        </Link>
        <Description>Заказ не найден</Description>
      </div>
    );
  }

  const number = order.number || order.id.slice(0, 8);
  const history = buildHistory(order);

  return (
    <div className={clsx(styles.root)}>
      <Link className={clsx(styles.backLink)} href={AppPath.CabinetOrders}>
        ← Мои заказы
      </Link>

      <header className={clsx(styles.header)}>
        <div className={clsx(styles.headerMain)}>
          <Typography.Heading className={clsx(styles.title)} level={1}>
            Заказ {number}
          </Typography.Heading>
          <p className={clsx(styles.subtitle)}>
            от {formatOrderDate(order.created_at)}
            {companyName ? ` · ${companyName}` : ''}
          </p>
        </div>

        <div className={clsx(styles.headerAside)}>
          <div className={clsx(styles.chips)}>
            <span className={clsx(listStyles.statusBadge, statusClassName(order.status))}>
              {formatOrderStatus(order.status)}
            </span>
            <span className={clsx(listStyles.statusBadge, paymentClassName(order.payment_status))}>
              {formatPaymentStatus(order.payment_status)}
            </span>
          </div>
          <div className={clsx(styles.actions)}>
            <Button className={clsx(styles.actionButton)} isDisabled variant="outline">
              Написать обращение
            </Button>
            <Button className={clsx(styles.actionButton)} isDisabled variant="outline">
              Повторить заказ
            </Button>
          </div>
        </div>
      </header>

      <div className={clsx(styles.layout)}>
        <section className={clsx(styles.card, styles.itemsCard)}>
          <h2 className={clsx(styles.cardTitle)}>Позиции</h2>
          {order.items.length === 0 ? (
            <Description>Состав заказа пуст</Description>
          ) : (
            <div className={clsx(styles.itemsWrap)}>
              <table className={clsx(styles.itemsTable)}>
                <thead>
                  <tr>
                    <th>Артикул</th>
                    <th>Наименование</th>
                    <th>Кол-во</th>
                    <th>Цена</th>
                    <th>Сумма</th>
                  </tr>
                </thead>
                <tbody>
                  {order.items.map((item) => (
                    <tr key={item.id}>
                      <td className={clsx(styles.itemSku)}>{item.sku || '—'}</td>
                      <td className={clsx(styles.itemName)}>{item.product_name || '—'}</td>
                      <td className={clsx(styles.itemQty)}>{item.qty}&nbsp;шт</td>
                      <td className={clsx(styles.itemPrice)}>{formatPrice(item.price)}</td>
                      <td className={clsx(styles.itemTotal)}>{formatPrice(item.total)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <div className={clsx(styles.totals)}>
                <div className={clsx(styles.totalRow)}>
                  <span>в т.ч. НДС 20%</span>
                  <strong>{formatPrice(order.vat)}</strong>
                </div>
                <div className={clsx(styles.totalRow, styles.totalStrong)}>
                  <span>Итого</span>
                  <strong>{formatPrice(order.total)}</strong>
                </div>
              </div>
            </div>
          )}
        </section>

        <aside className={clsx(styles.aside)}>
          <section className={clsx(styles.card)}>
            <h2 className={clsx(styles.cardTitle)}>Информация</h2>
            <dl className={clsx(styles.infoGrid)}>
              <div>
                <dt>Доставка</dt>
                <dd>{formatDeliveryType(order.delivery_type)}</dd>
              </div>
              <div>
                <dt>Адрес</dt>
                <dd>{displayValue(order.delivery_address)}</dd>
              </div>
              <div>
                <dt>Контакт</dt>
                <dd>
                  {displayValue(order.contact_name)}
                  {order.phone.trim() ? (
                    <span className={clsx(styles.infoPhone)}>{order.phone}</span>
                  ) : null}
                </dd>
              </div>
              <div>
                <dt>Ответственный отдел</dt>
                <dd>—</dd>
              </div>
              <div className={clsx(styles.infoWide)}>
                <dt>ERP</dt>
                <dd>—</dd>
              </div>
            </dl>
            <div className={clsx(styles.commentBlock)}>
              <span className={clsx(styles.commentLabel)}>Комментарий</span>
              <p className={clsx(styles.commentText)}>{displayValue(order.comment)}</p>
            </div>
          </section>

          <section className={clsx(styles.card)}>
            <h2 className={clsx(styles.cardTitle)}>Ваш отзыв</h2>
            <div aria-hidden className={clsx(styles.reviewEmpty)}>
              <span className={clsx(styles.reviewStars)}>★★★★★</span>
            </div>
            <p className={clsx(styles.reviewStub)}>Отзыв пока недоступен</p>
          </section>

          <section className={clsx(styles.card)}>
            <h2 className={clsx(styles.cardTitle)}>Документы</h2>
            {order.documents.length === 0 ? (
              <Description>Документов пока нет</Description>
            ) : (
              <ul className={clsx(styles.docsList)}>
                {order.documents.map((doc) => (
                  <li className={clsx(styles.docItem)} key={doc.id}>
                    <div className={clsx(styles.docMain)}>
                      <div className={clsx(styles.docTitleRow)}>
                        <span className={clsx(styles.docName)}>
                          {doc.name || doc.type || 'Документ'}
                        </span>
                        {doc.url ? <span className={clsx(styles.docBadge)}>Отправлен</span> : null}
                      </div>
                      <span className={clsx(styles.docMeta)}>
                        {formatOrderDate(doc.created_at)}
                      </span>
                    </div>
                    <div className={clsx(styles.docActions)}>
                      {doc.url ? (
                        <>
                          <a
                            className={clsx(styles.docLink)}
                            href={doc.url}
                            rel="noreferrer"
                            target="_blank"
                          >
                            Посмотреть
                          </a>
                          <a
                            className={clsx(styles.docLink)}
                            download
                            href={doc.url}
                            rel="noreferrer"
                          >
                            Скачать
                          </a>
                        </>
                      ) : (
                        <span className={clsx(styles.docMeta)}>Файл недоступен</span>
                      )}
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className={clsx(styles.card)}>
            <h2 className={clsx(styles.cardTitle)}>История</h2>
            {history.length === 0 ? (
              <Description>История пуста</Description>
            ) : (
              <ol className={clsx(styles.timeline)}>
                {history.map((item) => (
                  <li className={clsx(styles.timelineItem)} key={item.id}>
                    <span className={clsx(styles.timelineDot)} />
                    <div className={clsx(styles.timelineBody)}>
                      <span className={clsx(styles.timelineTitle)}>{historyLabel(item)}</span>
                      <span className={clsx(styles.timelineDate)}>
                        {formatOrderDateTime(item.created_at)}
                      </span>
                    </div>
                  </li>
                ))}
              </ol>
            )}
          </section>

          <CabinetOrderFeedback
            orderId={order.id}
            orderNumber={order.number || order.id}
            orderStatus={order.status}
          />
        </aside>
      </div>
    </div>
  );
};
