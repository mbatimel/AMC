'use client';

import { Button, FieldError, Form, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';

import { $authError, $isAuthPending, loginFx, sessionHydrated } from '@/core/entities/session';
import { IconLogin } from '@/core/shared/icons/IconLogin';
import { readFormString } from '@/core/shared/lib/readFormString';
import { validateEmail } from '@/core/shared/lib/validateContact';
import { AppPath } from '@/core/shared/router/paths';
import { resolveSafeNextPath } from '@/core/shared/router/resolveSafeNextPath';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

export const Login = (): JSX.Element => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [emailError, setEmailError] = useState<null | string>(null);
  const [authError, isPending, login, hydrate] = useUnit([
    $authError,
    $isAuthPending,
    loginFx,
    sessionHydrated,
  ]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const email = readFormString(formData, 'email').trim();
    const password = readFormString(formData, 'password');
    const nextEmailError = validateEmail(email);

    setEmailError(nextEmailError);

    if (nextEmailError) {
      return;
    }

    void login({ email, password })
      .then(() => {
        router.replace(resolveSafeNextPath(searchParams.get('next')));
      })
      .catch(() => {
        // error in $authError
      });
  };

  return (
    <AuthShell>
      <AuthCardHeader
        description="Оптовым клиентам после входа доступны индивидуальные цены, оформление и история заказов, документы."
        icon={IconLogin}
        title="Вход в личный кабинет"
      />
      <Form className={clsx(formStyles.form)} onSubmit={handleSubmit}>
        <TextField
          className={clsx(formStyles.field)}
          isInvalid={Boolean(emailError)}
          isRequired
          name="email"
          type="email"
        >
          <Label className={clsx(formStyles.label)}>E-mail</Label>
          <Input className={clsx(formStyles.input)} fullWidth placeholder="client@company.ru" />
          {emailError ? <FieldError>{emailError}</FieldError> : <FieldError />}
        </TextField>
        <TextField className={clsx(formStyles.field)} isRequired name="password" type="password">
          <Label className={clsx(formStyles.label)}>Пароль</Label>
          <Input className={clsx(formStyles.input)} fullWidth placeholder="••••••••" />
          <FieldError />
        </TextField>
        <div>
          <Link className={clsx(formStyles.forgotLink)} href={AppPath.ForgotPassword}>
            Забыли пароль?
          </Link>
        </div>
        {authError ? <p className={clsx(formStyles.error)}>{authError}</p> : null}
        <div className={clsx(formStyles.actions)}>
          <Button
            className={clsx(formStyles.submitButton)}
            isDisabled={isPending}
            type="submit"
            variant="primary"
          >
            <IconLogin currentColor="currentColor" height={16} width={16} />
            {isPending ? 'Вход…' : 'Войти'}
          </Button>
          <Link className={clsx(formStyles.secondaryLink)} href={AppPath.Home}>
            Продолжить как гость →
          </Link>
        </div>
      </Form>
      <div className={clsx(formStyles.footer)}>
        <span>Нет аккаунта?</span>
        <Link className={clsx(formStyles.footerLink)} href={AppPath.Register}>
          Зарегистрироваться
        </Link>
      </div>
    </AuthShell>
  );
};
