'use client';

import { Button, FieldError, Form, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useState } from 'react';

import { AuthApiError, requestPasswordResetRequest } from '@/core/shared/api/auth';
import { IconKey } from '@/core/shared/icons/IconKey';
import { IconMail } from '@/core/shared/icons/IconMail';
import { readFormString } from '@/core/shared/lib/readFormString';
import { validateEmail } from '@/core/shared/lib/validateContact';
import { AppPath } from '@/core/shared/router/paths';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

export const ForgotPassword = (): JSX.Element => {
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [emailSent, setEmailSent] = useState(true);
  const [emailError, setEmailError] = useState<null | string>(null);
  const [submitError, setSubmitError] = useState<null | string>(null);
  const [isPending, setIsPending] = useState(false);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const email = readFormString(formData, 'email').trim();
    const nextEmailError = validateEmail(email);

    setEmailError(nextEmailError);
    setSubmitError(null);

    if (nextEmailError) {
      return;
    }

    setIsPending(true);

    void requestPasswordResetRequest(email)
      .then((result) => {
        setEmailSent(result.emailSent);
        setIsSubmitted(true);
      })
      .catch((error: unknown) => {
        const message =
          error instanceof AuthApiError
            ? error.message
            : 'Не удалось отправить ссылку. Попробуйте позже.';

        setSubmitError(message);
      })
      .finally(() => {
        setIsPending(false);
      });
  };

  return (
    <AuthShell>
      <AuthCardHeader
        description="Укажите e-mail, привязанный к аккаунту — мы вышлем на него ссылку для сброса пароля."
        icon={IconKey}
        title="Восстановление пароля"
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
          <Input fullWidth placeholder="client@company.ru" />
          {emailError ? <FieldError>{emailError}</FieldError> : <FieldError />}
        </TextField>
        {isSubmitted ? (
          emailSent ? (
            <p className={clsx(formStyles.success)}>
              Если аккаунт с таким e-mail существует, ссылка для сброса пароля отправлена. Проверьте
              почту.
            </p>
          ) : (
            <p className={clsx(formStyles.error)}>
              Не удалось отправить письмо. Попробуйте ещё раз позже или обратитесь в поддержку.
            </p>
          )
        ) : null}
        {submitError ? <p className={clsx(formStyles.error)}>{submitError}</p> : null}
        <div className={clsx(formStyles.actions)}>
          <Button isDisabled={isPending} type="submit" variant="primary">
            <IconMail currentColor="currentColor" height={16} width={16} />
            {isPending ? 'Отправляем…' : 'Отправить ссылку'}
          </Button>
          <Link className={clsx(formStyles.secondaryLink)} href={AppPath.Login}>
            Назад ко входу
          </Link>
        </div>
      </Form>
    </AuthShell>
  );
};
