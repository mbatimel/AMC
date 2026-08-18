'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect, useState } from 'react';

import { SIGNUP_STATUS_LABELS } from '@/core/shared/api/signupRequests';

import styles from './Admin.module.css';
import { formatAdminDateTime } from './lib/nav';
import {
  $signupRequests,
  $usersError,
  adminUsersOpened,
  signupDecided,
  signupRejected,
} from './model/users';
import { AdminPageHeader } from './ui/AdminPageHeader';

export const AdminSignupRequestsPage = (): JSX.Element => {
  const [requests, error, open, decide, reject] = useUnit([
    $signupRequests,
    $usersError,
    adminUsersOpened,
    signupDecided,
    signupRejected,
  ]);
  const [rejectingId, setRejectingId] = useState('');
  const [rejectReason, setRejectReason] = useState('');

  useEffect(() => {
    open();
  }, [open]);

  const pending = requests.filter((request) => request.status === 'pending').length;

  return (
    <>
      <AdminPageHeader
        subtitle={`${pending} заявок ожидают решения`}
        title="Заявки на регистрацию"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <div className={clsx(styles.tableWrap)}>
        <table className={clsx(styles.table)}>
          <thead>
            <tr>
              <th>Дата</th>
              <th>Компания</th>
              <th>ИНН</th>
              <th>Контакт</th>
              <th>E-mail</th>
              <th>Статус</th>
              <th aria-label="Действия" />
            </tr>
          </thead>
          <tbody>
            {requests.length === 0 ? (
              <tr>
                <td className={clsx(styles.empty)} colSpan={7}>
                  Заявок пока нет
                </td>
              </tr>
            ) : null}

            {requests.map((request) => (
              <tr key={request.id}>
                <td>{formatAdminDateTime(request.created_at)}</td>
                <td>
                  <strong>{request.company || '—'}</strong>
                  <div className={clsx(styles.hint)}>
                    {request.type === 'organization' ? 'Организация / ИП' : 'Физическое лицо'}
                  </div>
                </td>
                <td>{request.inn || '—'}</td>
                <td>
                  {request.contact || '—'}
                  {request.phone ? <div className={clsx(styles.hint)}>{request.phone}</div> : null}
                </td>
                <td>{request.email}</td>
                <td>
                  <span
                    className={clsx(
                      styles.badge,
                      request.status === 'approved' && styles.badgeSuccess,
                      request.status === 'rejected' && styles.badgeAlert,
                      request.status === 'pending' && styles.badgeWarning,
                    )}
                  >
                    {SIGNUP_STATUS_LABELS[request.status]}
                  </span>
                  {request.reject_reason ? (
                    <div className={clsx(styles.hint)}>{request.reject_reason}</div>
                  ) : null}
                </td>
                <td>
                  {request.status === 'pending' ? (
                    <div className={clsx(styles.rowActions)}>
                      <button
                        className={clsx(styles.smallButton, styles.smallButtonPrimary)}
                        onClick={() => decide({ id: request.id, status: 'approved' })}
                        type="button"
                      >
                        Одобрить
                      </button>
                      <button
                        className={clsx(styles.smallButton, styles.smallButtonDanger)}
                        onClick={() => setRejectingId(request.id)}
                        type="button"
                      >
                        Отклонить
                      </button>
                    </div>
                  ) : (
                    <span className={clsx(styles.hint)}>—</span>
                  )}

                  {rejectingId === request.id ? (
                    <div className={clsx(styles.listEditor)}>
                      <input
                        aria-label="Причина отклонения"
                        className={clsx(styles.input)}
                        onChange={(event) => setRejectReason(event.target.value)}
                        placeholder="Причина отклонения"
                        value={rejectReason}
                      />
                      <div className={clsx(styles.rowActions)}>
                        <button
                          className={clsx(styles.smallButton, styles.smallButtonDanger)}
                          onClick={() => {
                            reject({ id: request.id, reason: rejectReason });
                            setRejectingId('');
                            setRejectReason('');
                          }}
                          type="button"
                        >
                          Подтвердить отказ
                        </button>
                        <button
                          className={clsx(styles.smallButton)}
                          onClick={() => setRejectingId('')}
                          type="button"
                        >
                          Отмена
                        </button>
                      </div>
                    </div>
                  ) : null}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
};
