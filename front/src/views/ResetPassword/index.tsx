'use client';

import { Button, FieldError, Form, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { useState } from 'react';

import { AuthApiError, confirmPasswordResetRequest } from '@/core/shared/api/auth';
import { IconKey } from '@/core/shared/icons/IconKey';
import { validatePassword, validateRequired } from '@/core/shared/lib/validateContact';
import { AppPath } from '@/core/shared/router/paths';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

export const ResetPassword = (): JSX.Element => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = (searchParams.get('token') ?? '').trim();

  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [submitError, setSubmitError] = useState<null | string>(null);
  const [isPending, setIsPending] = useState(false);

  const validateConfirmPassword = (value: string): null | string => {
    const requiredError = validateRequired(value);

    if (requiredError) {
      return requiredError;
    }

    if (value !== newPassword) {
      return 'Пароли не совпадают';
    }

    return null;
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    setSubmitError(null);

    if (!token) {
      setSubmitError('В ссылке нет токена сброса. Запросите новую ссылку.');

      return;
    }

    const passwordError = validatePassword(newPassword);
    const confirmError = validateConfirmPassword(confirmPassword);

    if (passwordError || confirmError) {
      return;
    }

    setIsPending(true);

    void confirmPasswordResetRequest({ newPassword, token })
      .then(() => {
        router.replace(AppPath.Login);
      })
      .catch((error: unknown) => {
        const message =
          error instanceof AuthApiError
            ? error.message
            : 'Не удалось сбросить пароль. Попробуйте позже.';

        setSubmitError(message);
      })
      .finally(() => {
        setIsPending(false);
      });
  };

  return (
    <AuthShell>
      <AuthCardHeader
        description="Придумайте новый пароль для входа в личный кабинет."
        icon={IconKey}
        title="Новый пароль"
      />

      {!token ? (
        <div className={clsx(formStyles.form)}>
          <p className={clsx(formStyles.error)}>
            Ссылка неполная или устарела. Запросите новую на странице восстановления пароля.
          </p>
          <div className={clsx(formStyles.actions)}>
            <Link className={clsx(formStyles.secondaryLink)} href={AppPath.ForgotPassword}>
              Запросить ссылку
            </Link>
            <Link className={clsx(formStyles.secondaryLink)} href={AppPath.Login}>
              Назад ко входу
            </Link>
          </div>
        </div>
      ) : (
        <Form className={clsx(formStyles.form)} onSubmit={handleSubmit}>
          <TextField
            className={clsx(formStyles.field)}
            isRequired
            minLength={6}
            name="newPassword"
            onChange={setNewPassword}
            type="password"
            validate={validatePassword}
            value={newPassword}
          >
            <Label className={clsx(formStyles.label)}>Новый пароль</Label>
            <Input fullWidth placeholder="••••••••" />
            <FieldError />
          </TextField>
          <TextField
            className={clsx(formStyles.field)}
            isRequired
            name="confirmPassword"
            onChange={setConfirmPassword}
            type="password"
            validate={validateConfirmPassword}
            value={confirmPassword}
          >
            <Label className={clsx(formStyles.label)}>Повторите пароль</Label>
            <Input fullWidth placeholder="••••••••" />
            <FieldError />
          </TextField>
          {submitError ? <p className={clsx(formStyles.error)}>{submitError}</p> : null}
          <div className={clsx(formStyles.actions)}>
            <Button isDisabled={isPending} type="submit" variant="primary">
              {isPending ? 'Сохраняем…' : 'Сохранить пароль'}
            </Button>
            <Link className={clsx(formStyles.secondaryLink)} href={AppPath.Login}>
              Назад ко входу
            </Link>
          </div>
        </Form>
      )}
    </AuthShell>
  );
};
