'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useEffect, useId, useState } from 'react';

import type { UserBlockPayload } from '@/core/shared/api/userBlock';

import {
  formatPhoneInput,
  normalizePhone,
  validateEmail,
  validatePhone,
} from '@/core/shared/lib/validateContact';

import styles from '../Admin.module.css';

type BlockUserDialogProps = {
  email: string;
  isPending?: boolean;
  onClose: () => void;
  onConfirm: (payload: UserBlockPayload) => void;
};

export const BlockUserDialog = ({
  email,
  isPending = false,
  onClose,
  onConfirm,
}: BlockUserDialogProps): JSX.Element => {
  const reasonId = useId();
  const contactNameId = useId();
  const contactPhoneId = useId();
  const contactEmailId = useId();
  const [reason, setReason] = useState('');
  const [contactName, setContactName] = useState('');
  const [contactPhone, setContactPhone] = useState('');
  const [contactEmail, setContactEmail] = useState('');
  const [phoneError, setPhoneError] = useState<null | string>(null);
  const [emailError, setEmailError] = useState<null | string>(null);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === 'Escape' && !isPending) {
        onClose();
      }
    };

    window.addEventListener('keydown', onKeyDown);

    return () => window.removeEventListener('keydown', onKeyDown);
  }, [isPending, onClose]);

  const handlePhoneChange = (value: string): void => {
    setContactPhone((previous) => formatPhoneInput(value, previous));
    setPhoneError(null);
  };

  const handleConfirm = (): void => {
    const nextPhoneError = validatePhone(contactPhone, { required: false });
    const nextEmailError = validateEmail(contactEmail, { required: false });

    setPhoneError(nextPhoneError);
    setEmailError(nextEmailError);

    if (nextPhoneError || nextEmailError) {
      return;
    }

    onConfirm({
      contactEmail: contactEmail.trim() || undefined,
      contactName: contactName.trim() || undefined,
      contactPhone: contactPhone.trim() ? normalizePhone(contactPhone) : undefined,
      reason: reason.trim() || undefined,
    });
  };

  return (
    <div
      aria-labelledby="block-user-dialog-title"
      aria-modal="true"
      className={clsx(styles.modalOverlay)}
      onClick={() => {
        if (!isPending) {
          onClose();
        }
      }}
      role="dialog"
    >
      <div className={clsx(styles.modalCard)} onClick={(event) => event.stopPropagation()}>
        <h2 className={clsx(styles.cardTitle)} id="block-user-dialog-title">
          Блокировка пользователя
        </h2>
        <p className={clsx(styles.hint)}>
          Пользователь <strong>{email}</strong> будет заблокирован. Все поля ниже необязательны.
        </p>
        <p className={clsx(styles.hint)}>
          Эта причина будет отправлена по email в уведомлении пользователю о блокировке. Контактное
          лицо и контакты, если заполнены, тоже попадут в уведомление и в сообщение при попытке
          входа.
        </p>

        <div className={clsx(styles.form)}>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor={reasonId}>
              Причина блокировки
            </label>
            <textarea
              className={clsx(styles.textarea)}
              id={reasonId}
              onChange={(event) => setReason(event.target.value)}
              placeholder="Например: нарушение условий работы"
              rows={3}
              value={reason}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor={contactNameId}>
              Контактное лицо
            </label>
            <input
              className={clsx(styles.input)}
              id={contactNameId}
              onChange={(event) => setContactName(event.target.value)}
              placeholder="ФИО менеджера"
              value={contactName}
            />
          </div>
          <div className={clsx(styles.formGrid)}>
            <div className={clsx(styles.field)}>
              <label className={clsx(styles.label)} htmlFor={contactPhoneId}>
                Телефон
              </label>
              <input
                aria-invalid={Boolean(phoneError)}
                className={clsx(styles.input)}
                id={contactPhoneId}
                inputMode="tel"
                onBlur={() => setPhoneError(validatePhone(contactPhone, { required: false }))}
                onChange={(event) => handlePhoneChange(event.target.value)}
                placeholder="+7 999 123 45 67"
                type="tel"
                value={contactPhone}
              />
              {phoneError ? <p className={clsx(styles.error)}>{phoneError}</p> : null}
            </div>
            <div className={clsx(styles.field)}>
              <label className={clsx(styles.label)} htmlFor={contactEmailId}>
                E-mail
              </label>
              <input
                aria-invalid={Boolean(emailError)}
                className={clsx(styles.input)}
                id={contactEmailId}
                onBlur={() => setEmailError(validateEmail(contactEmail, { required: false }))}
                onChange={(event) => {
                  setContactEmail(event.target.value);
                  setEmailError(null);
                }}
                placeholder="manager@company.ru"
                type="email"
                value={contactEmail}
              />
              {emailError ? <p className={clsx(styles.error)}>{emailError}</p> : null}
            </div>
          </div>
        </div>

        <div className={clsx(styles.actionsRow)}>
          <Button isDisabled={isPending} onPress={handleConfirm} variant="danger">
            {isPending ? 'Блокируем…' : 'Заблокировать'}
          </Button>
          <button
            className={clsx(styles.smallButton)}
            disabled={isPending}
            onClick={onClose}
            type="button"
          >
            Отмена
          </button>
        </div>
      </div>
    </div>
  );
};
