'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import styles from './Admin.module.css';
import { formatAdminDateTime } from './lib/nav';
import {
  $auditLog,
  $auditLogError,
  $auditLogTotal,
  $isAuditLogPending,
  adminAuditLogOpened,
} from './model/audit';
import { AdminPageHeader } from './ui/AdminPageHeader';

export const AdminAuditLogPage = (): JSX.Element => {
  const [entries, total, isPending, error, open] = useUnit([
    $auditLog,
    $auditLogTotal,
    $isAuditLogPending,
    $auditLogError,
    adminAuditLogOpened,
  ]);

  useEffect(() => {
    open();
  }, [open]);

  return (
    <>
      <AdminPageHeader
        subtitle={`${total} записей · хранятся без возможности удаления`}
        title="Журнал действий"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <div className={clsx(styles.tableWrap)}>
        <table className={clsx(styles.table)}>
          <thead>
            <tr>
              <th>Время</th>
              <th>Пользователь</th>
              <th>Действие</th>
            </tr>
          </thead>
          <tbody>
            {isPending && entries.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={3}>
                  Загружаем журнал…
                </td>
              </tr>
            ) : null}

            {!isPending && entries.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={3}>
                  Журнал пуст
                </td>
              </tr>
            ) : null}

            {entries.map((entry) => (
              <tr key={entry.id}>
                <td>{formatAdminDateTime(entry.createdAt)}</td>
                <td>{entry.actorLabel}</td>
                <td>{entry.action}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
};
