'use client';

import { Button, FieldError, Form, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useSearchParams } from 'next/navigation';
import { useEffect } from 'react';

import {
  $adminAuthError,
  $adminUserId,
  $isAdminAuthPending,
  adminLoginFx,
  adminSessionHydrated,
} from '@/core/entities/adminSession';
import { IconKey } from '@/core/shared/icons/IconKey';
import { readFormString } from '@/core/shared/lib/readFormString';
import { resolveSafeNextPath } from '@/core/shared/router/resolveSafeNextPath';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

export const AdminLoginPage = (): JSX.Element => {
  const searchParams = useSearchParams();
  const [authError, isPending, adminUserId, login, hydrate] = useUnit([
    $adminAuthError,
    $isAdminAuthPending,
    $adminUserId,
    adminLoginFx,
    adminSessionHydrated,
  ]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  useEffect(() => {
    if (!adminUserId) {
      return;
    }

    // Уже есть сессия (hydrate / cookie) — hard nav, чтобы proxy увидел cookie.
    window.location.assign(resolveSafeNextPath(searchParams.get('next')));
  }, [adminUserId, searchParams]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const nextPath = resolveSafeNextPath(searchParams.get('next'));

    void login({
      email: readFormString(formData, 'email').trim(),
      password: readFormString(formData, 'password'),
    })
      .then(() => {
        // Soft router.replace ломается: proxy не успевает / не видит cookie.
        window.location.assign(nextPath);
      })
      .catch(() => {
        // текст ошибки — в $adminAuthError
      });
  };

  return (
    <AuthShell>
      <AuthCardHeader
        description="Отдельный вход для сотрудников портала. Доступ выдаётся ролью «admin»."
        icon={IconKey}
        title="Вход в админ-панель"
      />
      <Form className={clsx(formStyles.form)} onSubmit={handleSubmit}>
        <TextField className={clsx(formStyles.field)} isRequired name="email" type="email">
          <Label className={clsx(formStyles.label)}>E-mail</Label>
          <Input className={clsx(formStyles.input)} fullWidth placeholder="admin@voint.ru" />
          <FieldError />
        </TextField>
        <TextField className={clsx(formStyles.field)} isRequired name="password" type="password">
          <Label className={clsx(formStyles.label)}>Пароль</Label>
          <Input className={clsx(formStyles.input)} fullWidth placeholder="••••••••" />
          <FieldError />
        </TextField>

        {authError ? <p className={clsx(formStyles.error)}>{authError}</p> : null}

        <div className={clsx(formStyles.actions)}>
          <Button
            className={clsx(formStyles.submitButton)}
            isDisabled={isPending}
            type="submit"
            variant="primary"
          >
            {isPending ? 'Вход…' : 'Войти'}
          </Button>
        </div>
      </Form>
    </AuthShell>
  );
};
