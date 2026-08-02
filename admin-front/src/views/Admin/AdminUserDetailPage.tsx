'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect } from 'react';

import { formatPrice } from '@/core/shared/lib/formatPrice';
import { AppPath } from '@/core/shared/router/paths';

import styles from './Admin.module.css';
import { formatAdminDateTime } from './lib/nav';
import {
  $isUserDetailPending,
  $isUserOrdersPending,
  $userDetail,
  $userDetailError,
  $userOrders,
  adminUserDetailOpened,
  adminUserDetailStatusToggled,
} from './model/userDetails';
import { AdminPageHeader } from './ui/AdminPageHeader';

const ORDER_STATUS_LABELS: Record<string, string> = {
  cancelled: 'Отменён',
  completed: 'Завершён',
  confirmed: 'Подтверждён',
  delivered: 'Доставлен',
  draft: 'Черновик',
  new: 'Новый',
  processing: 'На рассмотрении',
  shipped: 'Отгружен',
};

const PAYMENT_STATUS_LABELS: Record<string, string> = {
  not_paid: 'Не оплачен',
  paid: 'Оплачен',
  partially_paid: 'Частично оплачен',
  pending: 'Ожидает оплаты',
  unpaid: 'Не оплачен',
};

const formatOrderStatus = (status: string): string =>
  ORDER_STATUS_LABELS[status] ?? (status || '—');

const formatPaymentStatus = (status: string): string =>
  PAYMENT_STATUS_LABELS[status] ?? (status || '—');

type AdminUserDetailPageProps = {
  userId: string;
};

export const AdminUserDetailPage = ({ userId }: AdminUserDetailPageProps): JSX.Element => {
  const [user, orders, isUserPending, isOrdersPending, error, open, toggleStatus] = useUnit([
    $userDetail,
    $userOrders,
    $isUserDetailPending,
    $isUserOrdersPending,
    $userDetailError,
    adminUserDetailOpened,
    adminUserDetailStatusToggled,
  ]);

  useEffect(() => {
    open(userId);
  }, [open, userId]);

  return (
    <>
      <AdminPageHeader
        actions={
          <Link className={clsx(styles.smallButton)} href={AppPath.Users}>
            ← К списку пользователей
          </Link>
        }
        subtitle={user?.email}
        title="Профиль пользователя"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      {isUserPending && !user ? <p className={clsx(styles.hint)}>Загрузка…</p> : null}

      {user ? (
        <section className={clsx(styles.card)}>
          <h2 className={clsx(styles.cardTitle)}>Данные пользователя</h2>
          <div className={clsx(styles.grid2)}>
            <div>
              <p className={clsx(styles.hint)}>E-mail</p>
              <p>{user.email || '—'}</p>
            </div>
            <div>
              <p className={clsx(styles.hint)}>Телефон</p>
              <p>{user.phone || '—'}</p>
            </div>
            <div>
              <p className={clsx(styles.hint)}>ФИО</p>
              <p>
                {[user.last_name, user.first_name, user.middle_name].filter(Boolean).join(' ') ||
                  '—'}
              </p>
            </div>
            <div>
              <p className={clsx(styles.hint)}>Роль</p>
              <p>{user.role === 'admin' ? 'Администратор' : 'Клиент'}</p>
            </div>
            <div>
              <p className={clsx(styles.hint)}>Компания</p>
              <p>{user.company_name || '—'}</p>
            </div>
            <div>
              <p className={clsx(styles.hint)}>ИНН</p>
              <p>{user.inn || '—'}</p>
            </div>
            <div>
              <p className={clsx(styles.hint)}>Регистрация</p>
              <p>{formatAdminDateTime(user.created_at)}</p>
            </div>
            <div>
              <p className={clsx(styles.hint)}>Статус</p>
              <p>
                {user.is_active ? (
                  <span className={clsx(styles.badge, styles.badgeSuccess)}>Активен</span>
                ) : (
                  <span className={clsx(styles.badge, styles.badgeAlert)}>Заблокирован</span>
                )}
              </p>
            </div>
          </div>
          <div className={clsx(styles.actionsRow)}>
            <Button
              onPress={() => toggleStatus({ id: user.id, isActive: !user.is_active })}
              variant={user.is_active ? 'danger' : 'primary'}
            >
              {user.is_active ? 'Заблокировать' : 'Разблокировать'}
            </Button>
          </div>
        </section>
      ) : null}

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>История заказов</h2>

        {isOrdersPending ? <p className={clsx(styles.hint)}>Загрузка…</p> : null}

        <div className={clsx(styles.tableWrap)}>
          <table className={clsx(styles.table)}>
            <thead>
              <tr>
                <th>№ заказа</th>
                <th>Дата</th>
                <th>Статус</th>
                <th>Оплата</th>
                <th>Сумма</th>
              </tr>
            </thead>
            <tbody>
              {orders.length === 0 && !isOrdersPending ? (
                <tr>
                  <td className={clsx(styles.empty)} colSpan={5}>
                    Заказов пока нет.
                  </td>
                </tr>
              ) : null}

              {orders.map((order) => (
                <tr key={order.id}>
                  <td>{order.number || order.id}</td>
                  <td>{formatAdminDateTime(order.created_at)}</td>
                  <td>{formatOrderStatus(order.status)}</td>
                  <td>{formatPaymentStatus(order.payment_status)}</td>
                  <td>{formatPrice(order.total)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </>
  );
};
