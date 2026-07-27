'use client';

import { Button, FieldError, Form, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useState } from 'react';

import { IconKey } from '@/core/shared/icons/IconKey';
import { IconMail } from '@/core/shared/icons/IconMail';
import { AppPath } from '@/core/shared/router/paths';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

export const ForgotPassword = (): JSX.Element => {
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    setIsSubmitted(true);
  };

  return (
    <AuthShell>
      <AuthCardHeader
        description="Укажите e-mail, привязанный к аккаунту — мы вышлем на него ссылку для сброса пароля."
        icon={IconKey}
        title="Восстановление пароля"
      />
      <Form className={clsx(formStyles.form)} onSubmit={handleSubmit}>
        <TextField className={clsx(formStyles.field)} isRequired name="email" type="email">
          <Label className={clsx(formStyles.label)}>E-mail</Label>
          <Input fullWidth placeholder="client@company.ru" />
          <FieldError />
        </TextField>
        {isSubmitted ? (
          <p className={clsx(formStyles.success)}>
            Если аккаунт с таким e-mail существует, ссылка для сброса пароля будет отправлена.
          </p>
        ) : null}
        <div className={clsx(formStyles.actions)}>
          <Button type="submit" variant="primary">
            <IconMail currentColor="currentColor" height={16} width={16} />
            Отправить ссылку
          </Button>
          <Link className={clsx(formStyles.secondaryLink)} href={AppPath.Login}>
            Назад ко входу
          </Link>
        </div>
      </Form>
    </AuthShell>
  );
};
