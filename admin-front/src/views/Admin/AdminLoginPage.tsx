'use client';

import { Button, FieldError, Form, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
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
import { AppPath } from '@/core/shared/router/paths';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

export const AdminLoginPage = (): JSX.Element => {
  const router = useRouter();
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
    if (adminUserId) {
      router.replace(AppPath.Home);
    }
  }, [adminUserId, router]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);

    void login({
      email: readFormString(formData, 'email').trim(),
      password: readFormString(formData, 'password'),
    })
      .then(() => {
        router.replace(AppPath.Home);
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
          <Input className={clsx(formStyles.input)} fullWidth placeholder="admin@amc.ru" />
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
          <Link className={clsx(formStyles.secondaryLink)} href={AppPath.Home}>
            ← На сайт
          </Link>
        </div>
      </Form>
    </AuthShell>
  );
};
