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
  $inviteResult,
  $isInvitePending,
  $isTogglePending,
  $portalUsers,
  $usersError,
  adminUsersOpened,
  inviteResultDismissed,
  portalUserInvited,
  portalUserToggled,
} from './model/users';
import { AdminPageHeader } from './ui/AdminPageHeader';
import { BlockUserDialog } from './ui/BlockUserDialog';

export const AdminUsersPage = (): JSX.Element => {
  const [
    users,
    error,
    inviteResult,
    isInvitePending,
    isTogglePending,
    open,
    invite,
    dismissInviteResult,
    toggle,
  ] = useUnit([
    $portalUsers,
    $usersError,
    $inviteResult,
    $isInvitePending,
    $isTogglePending,
    adminUsersOpened,
    portalUserInvited,
    inviteResultDismissed,
    portalUserToggled,
  ]);
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteName, setInviteName] = useState('');
  const [copyState, setCopyState] = useState<'copied' | 'failed' | 'idle'>('idle');
  const [copiedPassword, setCopiedPassword] = useState<null | string>(null);
  const [blockTarget, setBlockTarget] = useState<null | { email: string; id: string }>(null);

  useEffect(() => {
    open();
  }, [open]);

  const passwordCopyState =
    inviteResult && copiedPassword === inviteResult.password ? copyState : 'idle';

  const admins = users.filter((user) => user.role === 'admin');
  const clients = users.filter((user) => user.role !== 'admin');

  const handleInvite = (): void => {
    const email = inviteEmail.trim();
    const name = inviteName.trim();

    if (!email || !name) {
      return;
    }

    invite({ email, name });
    setInviteEmail('');
    setInviteName('');
    setIsInviteOpen(false);
    setCopyState('idle');
    setCopiedPassword(null);
  };

  const handleCopyPassword = (): void => {
    if (!inviteResult?.password) {
      return;
    }

    const password = inviteResult.password;

    void navigator.clipboard.writeText(password).then(
      () => {
        setCopiedPassword(password);
        setCopyState('copied');
      },
      () => {
        setCopiedPassword(password);
        setCopyState('failed');
      },
    );
  };

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

      {inviteResult ? (
        <section className={clsx(styles.card)}>
          <h2 className={clsx(styles.cardTitle)}>Администратор создан</h2>
          <p className={clsx(styles.hint)}>Скопируйте пароль — он больше нигде не отобразится.</p>
          {!inviteResult.emailSent ? (
            <p className={clsx(styles.error)}>
              Письмо с доступом не удалось отправить. Передайте пароль администратору вручную.
            </p>
          ) : (
            <p className={clsx(styles.hint)}>Письмо с доступом отправлено на e-mail.</p>
          )}
          <div className={clsx(styles.formGrid)}>
            <div className={clsx(styles.field)}>
              <span className={clsx(styles.label)}>E-mail</span>
              <p>{inviteResult.email}</p>
            </div>
            <div className={clsx(styles.field)}>
              <span className={clsx(styles.label)}>Пароль</span>
              <code className={clsx(styles.invitePassword)}>{inviteResult.password}</code>
            </div>
          </div>
          <div className={clsx(styles.actionsRow)}>
            <Button onPress={handleCopyPassword} variant="primary">
              {passwordCopyState === 'copied' ? 'Скопировано' : 'Скопировать пароль'}
            </Button>
            <button
              className={clsx(styles.smallButton)}
              onClick={() => dismissInviteResult()}
              type="button"
            >
              Закрыть
            </button>
          </div>
          {passwordCopyState === 'failed' ? (
            <p className={clsx(styles.error)}>
              Не удалось скопировать — скопируйте пароль вручную.
            </p>
          ) : null}
        </section>
      ) : null}

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
              isDisabled={
                isInvitePending || inviteEmail.trim().length === 0 || inviteName.trim().length === 0
              }
              onPress={handleInvite}
              variant="primary"
            >
              {isInvitePending ? 'Отправляем…' : 'Отправить приглашение'}
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
            Создаётся учётная запись администратора. Пароль покажем один раз после создания.
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
                  Клиентов пока нет.
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
                      className={clsx(
                        styles.smallButton,
                        user.is_active && styles.smallButtonDanger,
                      )}
                      onClick={() => {
                        if (user.is_active) {
                          setBlockTarget({ email: user.email, id: user.id });

                          return;
                        }

                        toggle({ id: user.id, isActive: true });
                      }}
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

      {blockTarget ? (
        <BlockUserDialog
          email={blockTarget.email}
          isPending={isTogglePending}
          onClose={() => setBlockTarget(null)}
          onConfirm={(deactivate) => {
            toggle({ deactivate, id: blockTarget.id, isActive: false });
            setBlockTarget(null);
          }}
        />
      ) : null}
    </>
  );
};
