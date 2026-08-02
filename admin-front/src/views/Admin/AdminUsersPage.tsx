'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect, useState } from 'react';

import { getUserDetailPath } from '@/core/shared/router/paths';

import styles from './Admin.module.css';
import { formatAdminDateTime } from './lib/nav';
import {
  $portalUsers,
  $usersError,
  adminUsersOpened,
  portalUserInvited,
  portalUserPasswordReset,
  portalUserToggled,
} from './model/users';
import { AdminPageHeader } from './ui/AdminPageHeader';

export const AdminUsersPage = (): JSX.Element => {
  const [users, error, open, invite, toggle, resetPassword] = useUnit([
    $portalUsers,
    $usersError,
    adminUsersOpened,
    portalUserInvited,
    portalUserToggled,
    portalUserPasswordReset,
  ]);
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteName, setInviteName] = useState('');

  useEffect(() => {
    open();
  }, [open]);

  const admins = users.filter((user) => user.role === 'admin');
  const clients = users.filter((user) => user.role !== 'admin');

  return (
    <>
      <AdminPageHeader
        actions={
          <Button onPress={() => setIsInviteOpen((value) => !value)} variant="primary">
            Пригласить администратора
          </Button>
        }
        subtitle={`${clients.length} клиентов · ${admins.length} администраторов`}
        title="Пользователи портала"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      {isInviteOpen ? (
        <section className={clsx(styles.card)}>
          <h2 className={clsx(styles.cardTitle)}>Приглашение администратора</h2>
          <div className={clsx(styles.formGrid)}>
            <div className={clsx(styles.field)}>
              <label className={clsx(styles.label)} htmlFor="invite-email">
                E-mail
              </label>
              <input
                className={clsx(styles.input)}
                id="invite-email"
                onChange={(event) => setInviteEmail(event.target.value)}
                type="email"
                value={inviteEmail}
              />
            </div>
            <div className={clsx(styles.field)}>
              <label className={clsx(styles.label)} htmlFor="invite-name">
                Имя
              </label>
              <input
                className={clsx(styles.input)}
                id="invite-name"
                onChange={(event) => setInviteName(event.target.value)}
                value={inviteName}
              />
            </div>
          </div>
          <div className={clsx(styles.actionsRow)}>
            <Button
              isDisabled={inviteEmail.trim().length === 0}
              onPress={() => {
                invite({ company: inviteName.trim(), email: inviteEmail.trim() });
                setInviteEmail('');
                setInviteName('');
                setIsInviteOpen(false);
              }}
              variant="primary"
            >
              Отправить приглашение
            </Button>
            <button
              className={clsx(styles.smallButton)}
              onClick={() => setIsInviteOpen(false)}
              type="button"
            >
              Отмена
            </button>
          </div>
          <p className={clsx(styles.hint)}>
            Роль «admin» назначается в access-сервисе. Приглашение фиксируется в журнале действий.
          </p>
        </section>
      ) : null}

      {admins.length > 0 ? (
        <section className={clsx(styles.card)}>
          <h2 className={clsx(styles.cardTitle)}>Администраторы</h2>
          <p className={clsx(styles.hint)}>
            {admins.map((admin, index) => (
              <span key={admin.id}>
                {index > 0 ? ', ' : ''}
                <Link href={getUserDetailPath(admin.id)}>{admin.email}</Link>
              </span>
            ))}
          </p>
        </section>
      ) : null}

      <div className={clsx(styles.tableWrap)}>
        <table className={clsx(styles.table)}>
          <thead>
            <tr>
              <th>E-mail</th>
              <th>Компания</th>
              <th>ИНН</th>
              <th>Статус</th>
              <th>Регистрация</th>
              <th aria-label="Действия" />
            </tr>
          </thead>
          <tbody>
            {clients.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={6}>
                  Клиентов пока нет. Учётные записи создаются после одобрения заявок на регистрацию.
                </td>
              </tr>
            ) : null}

            {clients.map((user) => (
              <tr key={user.id}>
                <td>{user.email}</td>
                <td>
                  <strong>{user.company || '—'}</strong>
                  {user.contact ? <div className={clsx(styles.hint)}>{user.contact}</div> : null}
                </td>
                <td>{user.inn || '—'}</td>
                <td>
                  {user.is_active ? (
                    <span className={clsx(styles.badge, styles.badgeSuccess)}>Активен</span>
                  ) : (
                    <span className={clsx(styles.badge, styles.badgeAlert)}>Заблокирован</span>
                  )}
                </td>
                <td>{formatAdminDateTime(user.created_at)}</td>
                <td>
                  <div className={clsx(styles.rowActions)}>
                    <Link className={clsx(styles.smallButton)} href={getUserDetailPath(user.id)}>
                      Профиль
                    </Link>
                    <button
                      className={clsx(styles.smallButton)}
                      onClick={() => resetPassword(user.id)}
                      type="button"
                    >
                      Сбросить пароль
                    </button>
                    <button
                      className={clsx(
                        styles.smallButton,
                        user.is_active && styles.smallButtonDanger,
                      )}
                      onClick={() => toggle({ id: user.id, isActive: !user.is_active })}
                      type="button"
                    >
                      {user.is_active ? 'Заблокировать' : 'Разблокировать'}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
};
